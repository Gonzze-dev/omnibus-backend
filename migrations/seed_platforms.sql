-- Bays of TERMINAL DE GUALEGUAYCHÚ (created by seed-ar), A01..A12 from west to east.
-- A01 is at (-33.02086152608921, -58.53529535424005); each next bay is ~6 m east,
-- following the slight tilt of the building. Idempotent: skips the bays that already exist.
DO $$
DECLARE
    terminal_id UUID;
BEGIN
    SELECT uuid INTO terminal_id
    FROM bus_terminal
    WHERE postal_code = '2820' AND name = 'TERMINAL DE GUALEGUAYCHÚ';

    IF terminal_id IS NULL THEN
        RAISE EXCEPTION 'TERMINAL DE GUALEGUAYCHÚ not found: run seed-ar first';
    END IF;

    INSERT INTO platform (bay, coordinates, bus_terminal_id)
    SELECT 'A' || lpad(n::text, 2, '0'),
           jsonb_build_object(
               'lat', -33.02086152608921 + (n - 1) * 0.0000027,
               'lng', -58.53529535424005 + (n - 1) * 0.0000661
           ),
           terminal_id
    FROM generate_series(1, 12) AS n
    WHERE NOT EXISTS (
        SELECT 1 FROM platform p
        WHERE p.bus_terminal_id = terminal_id
          AND p.bay = 'A' || lpad(n::text, 2, '0')
    );
END $$;
