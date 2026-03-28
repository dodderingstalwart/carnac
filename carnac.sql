DROP TABLE IF EXISTS Insults;
CREATE TABLE Insults (ID int NOT NULL AUTO_INCREMENT,
                      insult varchar(255) NOT NULL,
                      PRIMARY KEY (ID)
);
DROP TABLE IF EXISTS Jokes;
CREATE TABLE Jokes (ID int NOT NULL AUTO_INCREMENT,
                    Answer varchar(255) NOT NULL,
                    Question varchar(255) NOT NULL,
                    PRIMARY KEY (ID)
);