package stream

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"

	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/master/pkg/stream"
)

const (
	// ProjectsDeleteKey specifies the key for delete projects.
	ProjectsDeleteKey = "projects_deleted"
	// ProjectsUpsertKey specifies the key for upsert projects.
	ProjectsUpsertKey = "project"
	// projectChannel specifies the channel to listen to project events.
	projectChannel = "stream_project_chan"
)

// JSONB is the golang equivalent of the postgres jsonb column type.
type JSONB interface{}

// ProjectMsg is a stream.Msg.
//
// determined:stream-gen source=server delete_msg=ProjectsDeleted
type ProjectMsg struct {
	bun.BaseModel `bun:"table:projects"`

	// immutable attributes
	ID int `bun:"id,pk" json:"id"`

	// mutable attributes
	Name        string      `bun:"name" json:"name"`
	Description string      `bun:"description" json:"description"`
	Archived    bool        `bun:"archived" json:"archived"`
	CreatedAt   time.Time   `bun:"created_at" json:"created_at"`
	Notes       JSONB       `bun:"notes" json:"notes"`
	WorkspaceID int         `bun:"workspace_id" json:"workspace_id"`
	UserID      int         `bun:"user_id" json:"user_id"`
	Immutable   bool        `bun:"immutable" json:"immutable"`
	State       model.State `bun:"state" json:"state"`

	// metadata
	Seq int64 `bun:"seq" json:"seq"`
}

// SeqNum gets the SeqNum from a ProjectMsg.
func (pm *ProjectMsg) SeqNum() int64 {
	return pm.Seq
}

// UpsertMsg creates a Project stream upsert message.
func (pm *ProjectMsg) UpsertMsg() stream.UpsertMsg {
	return stream.UpsertMsg{
		JSONKey: ProjectsUpsertKey,
		Msg:     pm,
	}
}

// DeleteMsg creates a Project stream delete message.
func (pm *ProjectMsg) DeleteMsg() stream.DeleteMsg {
	deleted := strconv.FormatInt(int64(pm.ID), 10)
	return stream.DeleteMsg{
		Key:     ProjectsDeleteKey,
		Deleted: deleted,
	}
}

// ProjectSubscriptionSpec is what a user submits to define a project subscription.
//
// determined:stream-gen source=client
type ProjectSubscriptionSpec struct {
	WorkspaceIDs []int `json:"workspace_ids"`
	ProjectIDs   []int `json:"project_ids"`
	Since        int64 `json:"since"`
}

// createFilteredProjectIDQuery creates a select query that
// pulls all relevant project ids based on permission scope and
// subscription spec filters.
func createFilteredProjectIDQuery(
	globalAccess bool,
	accessScopes []model.AccessScopeID,
	spec ProjectSubscriptionSpec,
) *bun.SelectQuery {
	q := db.Bun().NewSelect().
		TableExpr("projects p").
		Column("p.id").
		OrderExpr("p.id ASC")

	// add permission scope filter in event of non-global access
	if !globalAccess {
		q = permFilterQuery(q, accessScopes)
	}

	q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
		if len(spec.ProjectIDs) > 0 {
			q.WhereOr("p.id in (?)", bun.In(spec.ProjectIDs))
		}
		if len(spec.WorkspaceIDs) > 0 {
			q.WhereOr("p.workspace_id in (?)", bun.In(spec.WorkspaceIDs))
		}
		return q
	})
	return q
}

// ProjectCollectStartupMsgs collects ProjectMsg's that were missed prior to startup.
func ProjectCollectStartupMsgs(
	ctx context.Context,
	user model.User,
	known string,
	spec ProjectSubscriptionSpec,
) (
	[]stream.MarshallableMsg, error,
) {
	var out []stream.MarshallableMsg

	if len(spec.ProjectIDs) == 0 && len(spec.WorkspaceIDs) == 0 {
		// empty subscription: everything known should be returned as deleted
		out = append(out, stream.DeleteMsg{
			Key:     ProjectsDeleteKey,
			Deleted: known,
		})
		return out, nil
	}
	// step 0: get user's permitted access scopes
	accessMap, err := AuthZProvider.Get().GetProjectStreamableScopes(ctx, user)
	if err != nil {
		return nil, err
	}
	_, globalAccess := accessMap[model.GlobalAccessScopeID]
	var accessScopes []model.AccessScopeID
	// only populate accessScopes if user doesn't have global access
	if !globalAccess {
		for id, isPermitted := range accessMap {
			if isPermitted {
				accessScopes = append(accessScopes, id)
			}
		}
	}

	// step 1: calculate all ids matching this subscription
	oldEventsQuery := createFilteredProjectIDQuery(
		globalAccess,
		accessScopes,
		spec,
	)
	newEventsQuery := createFilteredProjectIDQuery(
		globalAccess,
		accessScopes,
		spec,
	)

	// get events that happened prior to since that are relevant (appearance)
	oldEventsQuery.Where("p.seq <= ?", spec.Since)
	var exist []int64
	err = oldEventsQuery.Scan(ctx, &exist)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		log.Errorf("error when scanning for old offline events: %v\n", err)
		return nil, err
	}
	// and events that happened since the last time this streamer checked
	newEventsQuery.Where("p.seq > ?", spec.Since)
	var newEntities []int64
	err = newEventsQuery.Scan(ctx, &newEntities)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		log.Errorf("error when scanning for new offline events: %v\n", err)
		return nil, err
	}

	exist = append(exist, newEntities...)
	slices.Sort(exist)

	// step 2: figure out what was missing and what has appeared
	missing, appeared, err := stream.ProcessKnown(known, exist)
	if err != nil {
		return nil, err
	}

	// step 3: hydrate appeared IDs into full ProjectMsgs
	var projMsgs []*ProjectMsg
	if len(appeared) > 0 {
		query := db.Bun().NewSelect().Model(&projMsgs).Where("project_msg.id in (?)", bun.In(appeared))
		if !globalAccess {
			query = permFilterQuery(query, accessScopes)
		}
		err := query.Scan(ctx, &projMsgs)
		if err != nil && errors.Cause(err) != sql.ErrNoRows {
			log.Errorf("error: %v\n", err)
			return nil, err
		}
	}

	// step 4: emit deletions and updates to the client
	out = append(out, stream.DeleteMsg{
		Key:     ProjectsDeleteKey,
		Deleted: missing,
	})
	for _, msg := range projMsgs {
		out = append(out, msg.UpsertMsg())
	}
	return out, nil
}

// ProjectMakeFilter creates a ProjectMsg filter based on the given ProjectSubscriptionSpec.
func ProjectMakeFilter(spec *ProjectSubscriptionSpec) (func(*ProjectMsg) bool, error) {
	// should this filter even run?
	if len(spec.WorkspaceIDs) == 0 && len(spec.ProjectIDs) == 0 {
		return nil, errors.Errorf("invalid subscription spec arguments: %v %v", spec.WorkspaceIDs, spec.ProjectIDs)
	}

	// create sets based on subscription spec
	workspaceIDs := make(map[int]struct{})
	for _, id := range spec.WorkspaceIDs {
		if id <= 0 {
			return nil, fmt.Errorf("invalid workspace id: %d", id)
		}
		workspaceIDs[id] = struct{}{}
	}
	projectIDs := make(map[int]struct{})
	for _, id := range spec.ProjectIDs {
		if id <= 0 {
			return nil, fmt.Errorf("invalid project id: %d", id)
		}
		projectIDs[id] = struct{}{}
	}

	// return a closure around our copied maps
	return func(msg *ProjectMsg) bool {
		// subscribed to project by this project_id?
		if _, ok := projectIDs[msg.ID]; ok {
			return true
		}
		// subscribed to this project by workspace_id?
		if _, ok := workspaceIDs[msg.WorkspaceID]; ok {
			return true
		}
		return false
	}, nil
}

// ProjectMakePermissionFilter returns a function that checks if a ProjectMsg
// is in scope of the user permissions.
func ProjectMakePermissionFilter(ctx context.Context, user model.User) (func(*ProjectMsg) bool, error) {
	accessScopeSet, err := AuthZProvider.Get().GetProjectStreamableScopes(ctx, user)
	if err != nil {
		return nil, err
	}

	switch {
	case accessScopeSet[model.GlobalAccessScopeID]:
		// user has global access for viewing projects
		return func(msg *ProjectMsg) bool { return true }, nil
	default:
		return func(msg *ProjectMsg) bool {
			return accessScopeSet[model.AccessScopeID(msg.WorkspaceID)]
		}, nil
	}
}
