CREATE TABLE IF NOT EXISTS sotrudniki(
    id          INT   PRIMARY KEY,
    name        TEXT  NOT NULL,
    secondname  TEXT  NOT NULL,
    job         TEXT  NOT NULL,
    otdel       INT   NOT NULL
);

INSERT INTO sotrudniki(id,name,secondname,job,otdel) VALUES
    (1,'Даниил','Дружинин','Разработчик',1),
    (2,'Кирилл','Михайлов','Разработчик',2),
    (3,'Даниил','Новиков','Тимлид',1),
    (4,'Алексей','Афанасьев','Архитектор',1),
    (5,'Никита','Авдеев','Дизайнер',3),
    (6,'Егор','Михайлов','Руководитель',4),
    (7,'Алексей','Михайлов','Веб дизайнер',2),
    (8,'Никита','Смирнов','Тестировщик',2),
    (9,'Богдан','Жданов','Тимлид',2),
    (10,'Андрей','Акимов','Аналитик',3)
    ON CONFLICT DO NOTHING;