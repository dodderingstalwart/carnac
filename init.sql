DROP TABLE IF EXISTS Insults; 
CREATE TABLE Insults (ID int NOT NULL AUTO_INCREMENT, 
                      Insult varchar(255) NOT NULL DISTINCT,
                      PRIMARY KEY (ID) 
);
CREATE TABLE Jokes (ID int NOT NULL AUTO_INCREMENT,
                    Answer varchar(255) NOT NULL,
                    Question varchar(255) NOT NULL,
                    PRIMARY KEY (ID)
);
