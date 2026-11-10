-- Params: columns, columns, rows, rows, start, start, end, end
-- Positions on the right and bottom edges (1) count towards the last
-- column and row.
SELECT
    MIN(CAST(x * ? AS INTEGER), ? - 1) AS col,
    MIN(CAST(y * ? AS INTEGER), ? - 1) AS row_index,
    COUNT(*) AS count
FROM cursors
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
GROUP BY col, row_index
