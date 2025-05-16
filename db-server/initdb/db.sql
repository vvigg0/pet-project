--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-05-16 15:24:21

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 218 (class 1259 OID 24706)
-- Name: sotrudniki; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sotrudniki (
    id integer NOT NULL,
    "имя" text NOT NULL,
    "фамилия" text NOT NULL,
    "должность" text NOT NULL,
    "отдел_id" integer NOT NULL
);


ALTER TABLE public.sotrudniki OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 24705)
-- Name: sotrudniki_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sotrudniki_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sotrudniki_id_seq OWNER TO postgres;

--
-- TOC entry 4897 (class 0 OID 0)
-- Dependencies: 217
-- Name: sotrudniki_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sotrudniki_id_seq OWNED BY public.sotrudniki.id;


--
-- TOC entry 4742 (class 2604 OID 24709)
-- Name: sotrudniki id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sotrudniki ALTER COLUMN id SET DEFAULT nextval('public.sotrudniki_id_seq'::regclass);


--
-- TOC entry 4891 (class 0 OID 24706)
-- Dependencies: 218
-- Data for Name: sotrudniki; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sotrudniki (id, "имя", "фамилия", "должность", "отдел_id") FROM stdin;
1	Иван	Иванов	Разработчик	1
2	Марина	Петрова	Тестировщик	2
3	Алексей	Сидоров	Бизнес-аналитик	1
4	Ольга	Кузнецова	Продакт-менеджер	3
5	Дмитрий	Смирнов	DevOps-инженер	2
6	Екатерина	Орлова	HR-специалист	4
7	Николай	Власов	Разработчик	1
8	Татьяна	Морозова	Технический писатель	3
9	Артур	Ким	Системный администратор	2
10	Анна	Гончарова	UI/UX-дизайнер	3
\.


--
-- TOC entry 4898 (class 0 OID 0)
-- Dependencies: 217
-- Name: sotrudniki_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sotrudniki_id_seq', 10, true);


--
-- TOC entry 4744 (class 2606 OID 24713)
-- Name: sotrudniki sotrudniki_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sotrudniki
    ADD CONSTRAINT sotrudniki_pkey PRIMARY KEY (id);


-- Completed on 2025-05-16 15:24:21

--
-- PostgreSQL database dump complete
--

