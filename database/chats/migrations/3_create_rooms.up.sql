
CREATE TABLE room (
  id UUID NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE (name)
);

CREATE TRIGGER trigger_room_updated_at
  BEFORE UPDATE OR INSERT ON room
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at();

CREATE TABLE room_user (
  room UUID NOT NULL,
  chat_user UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (room, chat_user),
  FOREIGN KEY (room) REFERENCES room(id),
  FOREIGN KEY (chat_user) REFERENCES chat_user(id)
);

CREATE INDEX room_user_chat_user_index ON room_user (chat_user);

CREATE TABLE user_ban (
  chat_user UUID NOT NULL,
  room UUID,
  valid_until TIMESTAMP WITH TIME ZONE,
  reason TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (chat_user),
  FOREIGN KEY (chat_user) REFERENCES chat_user(id),
  FOREIGN KEY (room) REFERENCES room(id),
  -- TODO: Migrating to psql 15 would allow NULLS NOT DISTINCT to make sure
  -- that we don't allow multiple productions from NULL buildings.
  -- See: https://stackoverflow.com/questions/8289100/create-unique-constraint-with-null-columns
  UNIQUE (chat_user, room)
);

CREATE INDEX user_ban_room_index ON user_ban (room);

CREATE TRIGGER trigger_user_ban_updated_at
  BEFORE UPDATE OR INSERT ON user_ban
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at();
