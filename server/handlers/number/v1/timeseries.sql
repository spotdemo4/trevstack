SELECT
    %s AS bucket,
    COUNT(*) AS count,
    COALESCE(SUM(number), 0) AS sum,
    COALESCE(AVG(number), 0.0) AS avg
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
GROUP BY bucket
ORDER BY bucket ASC
