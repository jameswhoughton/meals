CREATE TABLE planner (
	id INTEGER PRIMARY KEY,
	meal_id INTEGER NOT NULL,
	date DATE NOT NULL,
	UNIQUE(meal_id, date)
);
