DROP TABLE public.trial_snapshots;

ALTER TABLE public.trials ADD COLUMN restarts integer NOT NULL DEFAULT 0;

DROP TABLE public.runs;

DROP TYPE public.run_type;

CREATE TYPE public.task_type AS ENUM (
    'TRIAL'
);

CREATE TABLE public.task_runs (
    id integer NOT NULL,
    start_time timestamp without time zone NOT NULL DEFAULT now(),
    end_time timestamp without time zone NULL,
    task_type task_type NOT NULL,
    task_type_fk_id integer NOT NULL,
    CONSTRAINT trial_runs_id_trial_id_unique UNIQUE (task_type, task_type_fk_id, id)
);

