CREATE TABLE public.speisekarte (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  name character varying NOT NULL,
  description character varying,
  categories character varying[],
  tags character varying[],
  image character varying,
  price double precision NOT NULL DEFAULT 0,
  CONSTRAINT speisekarte_pkey PRIMARY KEY (id),
  CONSTRAINT speisekarte_id_key UNIQUE (id)
);
