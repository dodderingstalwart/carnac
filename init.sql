DROP TABLE IF EXISTS Insults; 
CREATE TABLE Insults (ID INTEGER PRIMARY KEY, 
                      Insult varchar(255) NOT NULL);

CREATE TABLE Jokes (ID INTEGER PRIMARY KEY,
                    Answer varchar(255) NOT NULL,
                    Question varchar(255) NOT NULL);
