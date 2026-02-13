-- Откат миграции: удаление таблиц в обратном порядке
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS users;
