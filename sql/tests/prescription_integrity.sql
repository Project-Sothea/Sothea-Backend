-- Run after the four setup files in a disposable database, with ON_ERROR_STOP=1.
-- All fixtures are rolled back.
BEGIN;
CREATE FUNCTION pg_temp.expect_error(command TEXT, expected TEXT) RETURNS VOID AS $$
BEGIN
  BEGIN
    EXECUTE command;
  EXCEPTION WHEN OTHERS THEN
    IF position(expected IN SQLERRM) = 0 THEN RAISE; END IF;
    RETURN;
  END;
  RAISE EXCEPTION 'Expected error containing % for %', expected, command;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
  p_patient_id INT := 900001;
  p_rx_id BIGINT;
  p_line_id BIGINT;
  p_drug_id BIGINT;
  p_batch_id BIGINT;
  p_location_id BIGINT;
  stock INT;
  packed BOOLEAN;
BEGIN
  INSERT INTO patient_details(id, name, family_group, khmer_name, dob, gender, village)
  VALUES(p_patient_id, 'Prescription regression', 'test', 'test', '2000-01-01', 'M', 'test');
  INSERT INTO admin(id, vid, reg_date, queue_no, pregnant, sent_to_id)
  VALUES(p_patient_id, 1, CURRENT_DATE, 'test', FALSE, FALSE),
        (p_patient_id, 2, CURRENT_DATE, 'test', FALSE, FALSE);
  INSERT INTO drugs(generic_name, dosage_form_code, route_code, strength_num, strength_unit_num, dispense_unit)
  VALUES('Prescription regression', 'TAB', 'PO', 500, 'mg', 'tab') RETURNING id INTO p_drug_id;
  INSERT INTO drug_batches(drug_id, batch_number) VALUES(p_drug_id, 'test') RETURNING id INTO p_batch_id;
  INSERT INTO batch_locations(batch_id, location, quantity) VALUES(p_batch_id, 'test', 100) RETURNING id INTO p_location_id;
  INSERT INTO prescriptions(patient_id, vid, notes) VALUES(p_patient_id, 1, 'original') RETURNING id INTO p_rx_id;
  INSERT INTO prescriptions(patient_id, vid, notes) VALUES(p_patient_id, 1, 'replacement')
  ON CONFLICT (patient_id, vid) DO NOTHING;
  IF (SELECT COUNT(*) FROM prescriptions WHERE prescriptions.patient_id = p_patient_id AND vid = 1) <> 1
     OR (SELECT notes FROM prescriptions WHERE id = p_rx_id) <> 'original' THEN
    RAISE EXCEPTION 'Duplicate creation changed the existing header';
  END IF;
  INSERT INTO prescription_lines(prescription_id, drug_id, dose_amount, dose_unit, frequency_code, duration, duration_unit)
  VALUES(p_rx_id, p_drug_id, 1, 'tab', 'OM', 10, 'day') RETURNING id INTO p_line_id;

  UPDATE prescription_lines SET is_packed = TRUE WHERE id = p_line_id;
  PERFORM pg_temp.expect_error('SET CONSTRAINTS ALL IMMEDIATE', 'Packed line must allocate exactly');
  -- Dispensing checks quantity even if an invalid packing update is still deferred.
  PERFORM pg_temp.expect_error(format('UPDATE prescriptions SET is_dispensed = TRUE WHERE id = %s', p_rx_id), 'exact allocations');
  UPDATE prescription_lines SET is_packed = FALSE WHERE id = p_line_id;
  SET CONSTRAINTS ALL IMMEDIATE;

  INSERT INTO prescription_batch_items(line_id, batch_location_id, quantity) VALUES(p_line_id, p_location_id, 10);
  UPDATE prescription_lines SET is_packed = TRUE WHERE id = p_line_id;
  DELETE FROM prescription_batch_items WHERE prescription_batch_items.line_id = p_line_id;
  SELECT is_packed INTO packed FROM prescription_lines WHERE id = p_line_id;
  SELECT quantity INTO stock FROM batch_locations WHERE id = p_location_id;
  IF packed OR stock <> 100 THEN RAISE EXCEPTION 'Clearing allocations did not unpack/return reserved stock'; END IF;
  INSERT INTO prescription_batch_items(line_id, batch_location_id, quantity) VALUES(p_line_id, p_location_id, 10);
  UPDATE prescription_lines SET is_packed = TRUE WHERE id = p_line_id;
  UPDATE prescription_batch_items SET quantity = 9 WHERE prescription_batch_items.line_id = p_line_id;
  IF (SELECT is_packed FROM prescription_lines WHERE id = p_line_id) THEN RAISE EXCEPTION 'Allocation update did not unpack'; END IF;
  UPDATE prescription_batch_items SET quantity = 10 WHERE prescription_batch_items.line_id = p_line_id;
  UPDATE prescription_lines SET is_packed = TRUE WHERE id = p_line_id;
  UPDATE prescription_lines SET duration = 9 WHERE id = p_line_id;
  IF (SELECT is_packed FROM prescription_lines WHERE id = p_line_id) THEN RAISE EXCEPTION 'Clinical change did not unpack'; END IF;
  UPDATE prescription_lines SET duration = 10 WHERE id = p_line_id;
  UPDATE prescription_lines SET is_packed = TRUE WHERE id = p_line_id;
  UPDATE prescriptions SET is_dispensed = TRUE WHERE id = p_rx_id;

  PERFORM pg_temp.expect_error(format('DELETE FROM admin WHERE id = %s AND vid = 1', p_patient_id), 'dispensing history');
  PERFORM pg_temp.expect_error(format('DELETE FROM patient_details WHERE id = %s', p_patient_id), 'dispensing history');
  PERFORM pg_temp.expect_error(format('DELETE FROM prescriptions WHERE id = %s', p_rx_id), 'dispensing history');
  PERFORM pg_temp.expect_error(format('DELETE FROM prescription_lines WHERE id = %s', p_line_id), 'dispensed prescription');
  PERFORM pg_temp.expect_error(format('DELETE FROM prescription_batch_items WHERE line_id = %s', p_line_id), 'dispensed prescription');
  SELECT quantity INTO stock FROM batch_locations WHERE id = p_location_id;
  IF stock <> 90 THEN RAISE EXCEPTION 'Dispensed stock returned: %', stock; END IF;
  PERFORM pg_temp.expect_error(format('UPDATE drugs SET strength_num = 250 WHERE id = %s', p_drug_id), 'Cannot modify strength');
  PERFORM pg_temp.expect_error(format('UPDATE drugs SET piece_content_amount = 50 WHERE id = %s', p_drug_id), 'Cannot modify strength');

  -- Cancelling an undispensed visit still releases its reservations.
  INSERT INTO prescriptions(patient_id, vid) VALUES(p_patient_id, 2) RETURNING id INTO p_rx_id;
  INSERT INTO prescription_lines(prescription_id, drug_id, dose_amount, dose_unit, frequency_code, duration, duration_unit)
  VALUES(p_rx_id, p_drug_id, 1, 'tab', 'OM', 7, 'day') RETURNING id INTO p_line_id;
  INSERT INTO prescription_batch_items(line_id, batch_location_id, quantity) VALUES(p_line_id, p_location_id, 7);
  DELETE FROM admin WHERE id = p_patient_id AND vid = 2;
  SELECT quantity INTO stock FROM batch_locations WHERE id = p_location_id;
  IF stock <> 90 THEN RAISE EXCEPTION 'Cancelled visit did not return reserved stock: %', stock; END IF;
END;
$$;
ROLLBACK;

