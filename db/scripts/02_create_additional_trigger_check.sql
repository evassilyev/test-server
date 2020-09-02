CREATE FUNCTION trigger_prevent_negative_balance()
    RETURNS trigger AS
$func$
BEGIN
    IF (select balance from calculated_balance_view) < 0 THEN
        RAISE EXCEPTION 'Attemp to change balance to negative';
    END IF;
    RETURN NULL;
END
$func$ LANGUAGE plpgsql;

CREATE TRIGGER negative_check_trg
    AFTER INSERT OR UPDATE ON balance_history
    FOR EACH STATEMENT
    EXECUTE PROCEDURE trigger_prevent_negative_balance();
