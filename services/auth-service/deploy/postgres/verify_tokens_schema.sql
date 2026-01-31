CREATE EXTENSION IF NOT EXISTS pgtap;

SELECT plan(1);

SELECT lives_ok('SELECT fn_verify_tokens_schema();', 'fn_verify_tokens_schema passes');

SELECT * FROM finish();
