
CREATE TABLE message (
  id UUID NOT NULL,
  chat_user UUID NOT NULL,
  room UUID NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (chat_user, room) REFERENCES room_user(chat_user, room),
  FOREIGN KEY (room) REFERENCES room(id)
);

CREATE INDEX message_chat_user_index ON message (chat_user);
CREATE INDEX message_room_index ON message (room);
CREATE INDEX message_created_at_index ON message (created_at);
