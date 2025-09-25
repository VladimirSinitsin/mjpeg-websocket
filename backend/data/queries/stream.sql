-- name: ListStreams :many
select s.id, s.title, s.description, s.frame_interval_ms, s.created_at, s.updated_at, count(f.id) as frame_count
from streams s left join frames f on f.stream_id = s.id
group by s.id
order by s.created_at desc
;

-- name: GetStream :one
select s.id, s.title, s.description, s.frame_interval_ms, s.created_at, s.updated_at, count(f.id) as frame_count
from streams s left join frames f on f.stream_id = s.id
group by s.id
having s.id = $1
;

-- name: UpdateStream :one
UPDATE streams s
SET
    updated_at = now(),
    title = $2,
    description = $3,
    frame_interval_ms = $4
WHERE s.id = $1
    RETURNING
    s.id,
    s.title,
    s.description,
    s.frame_interval_ms,
    s.created_at,
    s.updated_at,
    (
        SELECT count(f.id)
        FROM frames f
        WHERE f.stream_id = s.id
    ) AS frame_count
;
