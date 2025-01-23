-- user1
INSERT INTO chat_server_schema.chat_user ("id", "name", "api_user")
  VALUES (
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11',
    'user1',
    '0463ed3d-bfc9-4c10-b6ee-c223bbca0fab'
  );

-- another-test-user@another-provider.com
INSERT INTO chat_server_schema.chat_user ("id", "name", "api_user")
  VALUES (
    '6e30c4e4-3906-4f7d-9919-fc2539b14301',
    'some-dude',
    '4f26321f-d0ea-46a3-83dd-6aa1c6053aaf'
  );

-- better-test-user@mail-client.org
INSERT INTO chat_server_schema.chat_user ("id", "name", "api_user")
  VALUES (
    'd9bf6974-3f1c-4fb8-a389-e75e0ed2b9bf',
    'i-m-banned',
    '00b265e6-6638-4b1b-aeac-5898c7307eb8'
  );

-- room-1
INSERT INTO chat_server_schema.room ("id", "name")
  VALUES (
    'ef3cc94b-5142-4399-a366-70645a504219',
    'room-1'
  );

INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'ef3cc94b-5142-4399-a366-70645a504219',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11'
  );

INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    'fa5e2b80-1c1e-417a-8195-21cb5c60c4ae',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11',
    'ef3cc94b-5142-4399-a366-70645a504219',
    'This is empty here'
  );

-- room-with-banned-people
INSERT INTO chat_server_schema.room ("id", "name")
  VALUES (
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'room-wit-banned-people'
  );

INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11'
  );

INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    '6e30c4e4-3906-4f7d-9919-fc2539b14301'
  );

INSERT INTO chat_server_schema.room_user ("room", "chat_user")
  VALUES (
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'd9bf6974-3f1c-4fb8-a389-e75e0ed2b9bf'
  );

INSERT INTO chat_server_schema.user_ban ("chat_user", "room", "valid_until", "reason")
  VALUES (
    'd9bf6974-3f1c-4fb8-a389-e75e0ed2b9bf',
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    current_timestamp + make_interval(hours => 6),
    'i do not like you'
  );

INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    '996c458c-1ec6-4b25-8c93-d3db9e805675',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11',
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'Hello'
  );
INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    'ecdeedde-50e2-4e26-81bd-7d651e667d17',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11',
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'Test message'
  );
INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    '6ea3fde2-113a-4b8f-b2d5-c118b40514c1',
    '6e30c4e4-3906-4f7d-9919-fc2539b14301',
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'Yes it works'
  );
INSERT INTO chat_server_schema.message ("id", "chat_user", "room", "message")
  VALUES (
    '3a44e2ee-65c6-4d80-baa2-2e7cbe32e5aa',
    '0198ed26-8e92-4b81-aec0-aaaff33b6a11',
    'b9d66811-c6f2-4f20-9374-4fc754c5098f',
    'Great thanks!'
  );

-- room-with-nobody
INSERT INTO chat_server_schema.room ("id", "name")
  VALUES (
    '7c3515ed-3c23-4761-a822-b004814b8e73',
    'room-with-nobody'
  );
