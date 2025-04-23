CREATE TABLE ingredients (
	id INTEGER PRIMARY KEY,
	user_id INTEGER,
	name TEXT NOT NULL,
	UNIQUE(user_id, name),
	FOREIGN KEY(user_id) REFERENCES users(id)
);
