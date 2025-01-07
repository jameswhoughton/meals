CREATE TABLE meals (
	id INT PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	notes TEXT NULL,
	quick BIT DEFAULT 0 NOT NULL,
	family BIT DEFAULT 0 NOT NULL,
	easy BIT DEFAULT 0 NOT NULL,
	main_ingredient_id INT NOT NULL,
	min_frequency INT  NOT NULL,
	last_eaten_at DATE
);
