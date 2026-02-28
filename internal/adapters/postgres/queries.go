package postgres

// message queries
const (
	saveMessageQuery = `
		INSERT INTO core.messages (sender_id, sender_role, room_id, content, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_id, sender_role, room_id, content, tags, created_at;
	`
	getLastMessagesQuery = `
		SELECT id, sender_id, sender_role, room_id, content, tags, created_at
		FROM core.messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2;
	`
)

// room queries
const (
	createRoomQuery = `
		INSERT INTO core.rooms (human_id)
		VALUES ($1)
		RETURNING id, human_id, ai_id, status, created_at, closed_at;
	`
	claimRoomQuery = `
		UPDATE core.rooms
		SET ai_id = $1, status = 'active'
		WHERE id = $2 AND ai_id IS NULL
		RETURNING id;
	`
	checkRoomQuery = `
		SELECT status,
		       (human_id = $2 OR ai_id = $2) AS is_participant
		FROM core.rooms
		WHERE id = $1;
	`
	closeRoomQuery = `
		UPDATE core.rooms
		SET status = 'closed', closed_at = NOW()
		WHERE id = $1
		  AND (human_id = $2 OR ai_id = $2)
		  AND status IN ('open', 'active');
	`
	getRoomQuery = `
		SELECT id, human_id, ai_id, status, created_at, closed_at
		FROM core.rooms
		WHERE id = $1;
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
