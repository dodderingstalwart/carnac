DROP TABLE IF EXISTS Insults; 
CREATE TABLE Insults (ID int NOT NULL AUTO_INCREMENTS, 
                      Insult varchar(255) NOT NULL,
                      PRIMARY KEY (ID) 
);
CREATE TABLE Jokes (ID int NOT NULL AUTO_INCREMENTS,
                    Answer varchar(255) NOT NULL,
                    Question varchar(255) NOT NULL,
                    PRIMARY KEY (ID)
);
