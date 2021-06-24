DROP TABLE public.trial_snapshots;

ALTER TABLE public.trials ADD COLUMN restarts integer NOT NULL DEFAULT 0;

DROP TABLE public.runs;

DROP TYPE public.run_type;

CREATE TYPE public.task_type AS ENUM (
    'TRIAL'
);

-- Task runs represent the multiple runs of a task, e.g. in the event
-- a task fails and is restarted, there may be multiple task runs for
-- a single task.
CREATE TABLE public.task_runs (
    id SERIAL,
    run_id integer NOT NULL,
    start_time timestamp without time zone NOT NULL DEFAULT now(),
    end_time timestamp without time zone NULL,
    -- These would be dropped when the below table exists.
    task_type task_type NOT NULL,
    task_type_fk_id integer NOT NULL,
    CONSTRAINT task_runs_id_trial_id_unique UNIQUE (task_type, task_type_fk_id, id)
);

-- This table is here solely to color how the task table may look.
CREATE TABLE public.tasks (
    id SERIAL,
    task_id uuid NOT NULL,
    start_time timestamp without time zone NOT NULL DEFAULT now(),
    end_time timestamp without time zone NULL,
    task_type task_type NOT NULL,
    task_type_fk_id integer NOT NULL,
    CONSTRAINT task_runs_id_trial_id_unique UNIQUE (task_type, task_type_fk_id, task_id)
);

CREATE TABLE public.trial_runner_metadata (
    id SERIAL,
    trial_id int,
    state text,
    CONSTRAINT trial_runner_metadata_trial_id_unique UNIQUE (trial_id),
    CONSTRAINT trial_runner_metadata_trial_id_fk FOREIGN KEY (trial_id) REFERENCES public.trials(id)
);
