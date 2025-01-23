
-- https://stackoverflow.com/questions/72985242/securely-create-role-in-postgres-using-sql-script-and-environment-variables
CREATE USER chat_server_admin WITH CREATEDB PASSWORD :'admin_password';
CREATE USER chat_server_manager WITH PASSWORD :'manager_password';
CREATE USER chat_server_user WITH PASSWORD :'user_password';

GRANT chat_server_user TO chat_server_manager;
GRANT chat_server_manager TO chat_server_admin;
