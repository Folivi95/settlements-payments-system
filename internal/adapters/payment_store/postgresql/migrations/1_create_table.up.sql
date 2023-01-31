CREATE TABLE IF NOT EXISTS payment_instructions (
    payment_instruction_id varchar(100) primary key,
    body jsonb not null default '{}'::jsonb
);