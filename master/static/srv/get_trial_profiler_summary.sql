SELECT
--         labels -> 'trialId' as trialId,
        *
FROM trial_profiler_metrics
WHERE labels @> '{"trialId":5,"name":"net_throughput_recv"}'::jsonb
ORDER BY batches[0]
LIMIT 100;

-- SELECT
--     array_to_json(m.values) AS values,
--     array_to_json(m.batches) AS batches,
--     array_to_json(m.ts) AS timestamps,
--     m.labels AS labels
-- FROM trial_profiler_metrics m
-- WHERE m.labels @> $1::jsonb
-- ORDER by m.id
-- OFFSET $2 LIMIT $3;
