CREATE TABLE meals (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	notes TEXT NULL,
	quick BIT DEFAULT 0 NOT NULL,
	family BIT DEFAULT 0 NOT NULL,
	easy BIT DEFAULT 0 NOT NULL,
	main_ingredient_id INTEGER NOT NULL,
	min_frequency INTEGER  NOT NULL,
	last_eaten_at DATE
);
