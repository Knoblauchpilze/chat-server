-- main room where everybody is added by default
INSERT INTO chat_server_schema.room ("id", "name")
  VALUES (
    'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68',
    'general'
  );

-- system user to communicate important updates
INSERT INTO chat_server_schema.chat_user ("id", "name", "api_user")
  VALUES (
    '0edebba5-fd0b-433d-bb6b-35ceb9fcb9b3',
    'chatterly-bot',
    '56a916a1-6e19-4e6d-aa38-adc22d6b049b'
  );

-- register the system user in the general room
INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68',
    '0edebba5-fd0b-433d-bb6b-35ceb9fcb9b3'
  );

-- post a welcome message
INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    'd0080da7-e95a-4ba3-ad0d-78ed72688e2e',
    '0edebba5-fd0b-433d-bb6b-35ceb9fcb9b3',
    'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68',
    'Welcome to chatterly!'
  );

-- ghost user to assign messages of deleted users to
INSERT INTO chat_server_schema.chat_user ("id", "name", "api_user")
  VALUES (
    '62a23b20-d28b-427f-85b0-db8f8feed0be',
    'ghost',
    '0ac3a469-2257-4989-9a4c-5539cd004776'
  );

-- register the system user in the general room
INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68',
    '62a23b20-d28b-427f-85b0-db8f8feed0be'
  );
