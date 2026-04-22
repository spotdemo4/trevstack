SELECT
    rowid,
    timestamp,
    name,
    number
FROM numbers
WHERE
    (? IS NULL OR name LIKE '%' || ? || '%')
    AND (? IS NULL OR number >= ?)
    AND (? IS NULL OR number <= ?)
    AND (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
    AND (? IS NULL OR rowid < ?)
ORDER BY rowid DESC
LIMIT ?
