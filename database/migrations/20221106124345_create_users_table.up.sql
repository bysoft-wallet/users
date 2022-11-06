CREATE TABLE public.users (
	uuid uuid NULL,
	email varchar NOT NULL,
	name varchar NOT NULL,
	created_at timestamp NULL,
	updated_at timestamp NULL,
	CONSTRAINT users_pk PRIMARY KEY (uuid)
);
