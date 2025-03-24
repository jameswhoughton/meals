CREATE TABLE schedule (
	meal_id INTEGER NOT NULL,
	date DATE NOT NULL,
	UNIQUE(meal_id, date)
);
