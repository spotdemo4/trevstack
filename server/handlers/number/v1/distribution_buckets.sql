-- Params: bucketCount-1, lo, bucketCount, span, start, start, end, end
-- bucket index = min(bucketCount-1, (number - lo) * bucketCount / span)
SELECT
    CAST(MIN(?, (number - ?) * ? / ?) AS INTEGER) AS idx,
    COUNT(*) AS bucket_count
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
GROUP BY idx
ORDER BY idx
