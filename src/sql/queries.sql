CREATE TABLE users.users(userid UUID PRIMARY KEY,
                                             name TEXT NOT NULL,
                                                       email TEXT NOT NULL,
                                                                  password TEXT NOT NULL,
                                                                                createdAt TIMESTAMP NOT NULL,
                                                                                                    updatedAt TIMESTAMP NOT NULL);


select *
from users.users
INSERT INTO users.users
VALUES ('b2a0b669-ebf4-433f-8f59-898a79f61d4e',
        'Bon Jovi',
        'bonjovi@email.com',
        'passwordbonjovi',
        current_timestamp,
        current_timestamp);


INSERT INTO users.users
VALUES ('b2a0b669-ebf4-433f-8f59-898a79f61d4f',
        'Rick Ashtley',
        'rickastley@email.com',
        'passwordrickashtley',
        current_timestamp,
        current_timestamp);


CREATE TABLE users.tasks(taskid UUID PRIMARY KEY,
                                             description TEXT NOT NULL,
                                                              createdat TIMESTAMP NOT NULL,
                                                                                  updatedat TIMESTAMP NOT NULL,
                                                                                                      validtill TIMESTAMP,
                                                                                                                userid UUID,
                                                                                                                CONSTRAINT fk_users
                         FOREIGN KEY (userid) REFERENCES users.users(userid))
select *
from users.tasks
INSERT INTO users.tasks
VALUES ('91f6cb09-b48c-42f8-a143-6cdabd1a1284',
        'First dummy task',
        current_timestamp,
        current_timestamp,
        'b2a0b669-ebf4-433f-8f59-898a79f61d4e');


CREATE TABLE users.session(sessionid UUID PRIMARY KEY,
                                                  userid UUID,
                                                  validtill TIMESTAMP,
                                                            CONSTRAINT fk_users
                           FOREIGN KEY (userid) REFERENCES users.users(userid))