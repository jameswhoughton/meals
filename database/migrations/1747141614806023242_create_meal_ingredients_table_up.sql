CREATE TABLE meal_ingredients (
	id INT NOT NULL AUTO_INCREMENT,
	meal_id INT NOT NULL,
	name VARCHAR(255),
	quantity INT,
	unit VARCHAR(100),
	PRIMARY KEY (id),
	FOREIGN KEY (meal_id) REFERENCES meals (id) ON DELETE CASCADE
);
