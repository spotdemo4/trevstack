SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(number), 0) AS total_sum,
    COALESCE(AVG(number), 0.0) AS average,
    COALESCE(MIN(number), 0) AS min_number,
    COALESCE(MAX(number), 0) AS max_number,
    COUNT(DISTINCT name) AS distinct_names
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
