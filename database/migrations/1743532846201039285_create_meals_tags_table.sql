CREATE TABLE meals_tags (
	meal_id INTEGER,
	tag_id INTEGER,
	UNIQUE(meal_id, tag_id)
);
