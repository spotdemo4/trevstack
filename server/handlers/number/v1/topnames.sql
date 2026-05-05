SELECT
    name,
    COUNT(*) AS total_count,
    COALESCE(SUM(number), 0) AS total_sum,
    COALESCE(AVG(number), 0.0) AS average
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
GROUP BY name
ORDER BY total_count DESC, total_sum DESC
LIMIT ?
