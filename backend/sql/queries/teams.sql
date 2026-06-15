-- name: CreateTeam :one
INSERT INTO teams (id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1;

-- name: AddTeamMember :one
INSERT INTO team_members (team_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMembership :one
SELECT * FROM team_members
WHERE team_id = $1 AND user_id = $2;

-- name: ListMembersByTeam :many
SELECT * FROM team_members
WHERE team_id = $1
ORDER BY created_at;

-- name: ListTeamsForUser :many
SELECT teams.*
FROM teams
JOIN team_members ON team_members.team_id = teams.id
WHERE team_members.user_id = $1
ORDER BY teams.created_at;

-- name: UpdateMemberRole :one
UPDATE team_members
SET role = $3
WHERE team_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveTeamMember :exec
DELETE FROM team_members
WHERE team_id = $1 AND user_id = $2;

-- name: CountTeamAdmins :one
SELECT count(*) FROM team_members
WHERE team_id = $1 AND role = 'admin';

-- name: ListTeamsByIDs :many
SELECT * FROM teams
WHERE id = ANY(sqlc.arg(ids)::uuid[]);

-- name: ListMembersByTeamIDs :many
SELECT * FROM team_members
WHERE team_id = ANY(sqlc.arg(team_ids)::uuid[])
ORDER BY team_id, created_at;
