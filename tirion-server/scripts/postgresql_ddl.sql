/* Settings */


/* Drops */

DROP TABLE IF EXISTS run;

/* Tables */

CREATE TABLE run (
	id SERIAL,
	name TEXT NOT NULL,
	sub_name TEXT NOT NULL,
	interval INT NOT NULL,
	metrics TEXT NOT NULL,
	prog TEXT NOT NULL,
	prog_arguments TEXT NOT NULL,
	start TIMESTAMP NOT NULL,
	stop TIMESTAMP,
	PRIMARY KEY(id)
);

/* new Settings */

/* Foreign Keys */

/* Indizes */

CREATE INDEX run_name_idx ON run(name);
