package message

const (
	saveMessageQuery = `
		INSERT INTO core.messages (id, sender_id, sender_role, room_id, content)
		VALUES (:id, :sender_id, :sender_role, :room_id, :content)
	`
	getLastMessagesQuery = `
		SELECT id, sender_id, sender_role, room_id, content, created_at
		FROM core.messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	getUserRoleQuery = `
		SELECT role FROM core.profiles WHERE user_id = $1
	`
)
