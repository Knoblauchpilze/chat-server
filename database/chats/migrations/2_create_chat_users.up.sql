
CREATE TABLE chat_user (
  id UUID NOT NULL,
  name TEXT NOT NULL,
  api_user UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  version INTEGER DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE (name)
);

CREATE TRIGGER trigger_chat_user_updated_at
  BEFORE UPDATE OR INSERT ON chat_user
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at();

CREATE INDEX chat_user_name_index ON chat_user (name);
