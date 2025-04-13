
DELETE FROM message WHERE id = 'd0080da7-e95a-4ba3-ad0d-78ed72688e2e';
DELETE FROM room_user WHERE
  room = 'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68'
  AND chat_user = '0edebba5-fd0b-433d-bb6b-35ceb9fcb9b3';
DELETE FROM chat_user WHERE id = '0edebba5-fd0b-433d-bb6b-35ceb9fcb9b3';
DELETE FROM room WHERE id = 'b2c0d9c8-c5bd-42ea-88e2-15b66fffdd68';
