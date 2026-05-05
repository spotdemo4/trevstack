SELECT
    COALESCE(MIN(number), 0) AS lo,
    COALESCE(MAX(number), 0) AS hi,
    COUNT(*) AS total
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
