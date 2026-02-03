package postgres

// message queries
const (
	saveMessageQuery = `
		INSERT INTO core.messages (sender_id, sender_role, room_id, content, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_id, sender_role, room_id, content, tags, assigned_to, created_at;
	`
	getLastMessagesQuery = `
		SELECT id, sender_id, sender_role, room_id, content, tags, created_at
		FROM core.messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2;
	`
	claimMessageQuery = `
		UPDATE core.messages SET assigned_to = $1 WHERE id = $2;
	`
	getMessageAssigneeQuery = `
		SELECT assigned_to FROM core.messages WHERE id = $1;
	`
)

// tag queries
const (
	getTagsQuery = `
		SELECT name FROM core.tags;
	`
	deleteProfileTagsQuery = `
		DELETE FROM core.profile_tags WHERE user_id = $1 RETURNING tag_name;
	`
	insertProfileTagsQuery = `
		INSERT INTO core.profile_tags (user_id, tag_name)
		VALUES ($1, $2);
	`
	getProfileTagsQuery = `
		SELECT tag_name FROM core.profile_tags WHERE user_id = $1;
	`
)
