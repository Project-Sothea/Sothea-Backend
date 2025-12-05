# Drug Presentation Unit Combinations & Prescription Units

## Unit Types
- **Mass units**: `mcg`, `mg`, `g`, `IU`
- **Volume units**: `mL`, `L`
- **Piece units**: `tab`, `cap`, `drop`, `bottle`

---

## Case 1: Solid with Known Strength
**Strength**: `strength_num` + `strength_unit_num` (mass/IU per piece)  
**Strength Den**: NULL  
**Dispense Unit**: `tab`, `cap`, or `drop`  
**Piece Content**: NULL

**Example**: Paracetamol 500 mg tablet
- Strength: `500 mg`
- Dispense Unit: `tab`

### Acceptable Prescription Units (dose_unit):
1. **`dispense_unit`** (e.g., `tab`, `cap`, `drop`) - Direct piece dispensing
2. **`strength_unit_num`** (e.g., `mg`, `g`, `IU`) - Convert using strength

**Example Prescriptions**:
- `dose_unit = 'tab'`, `dose_amount = 2` → Dispense 2 tablets
- `dose_unit = 'mg'`, `dose_amount = 1000` → Dispense 2 tablets (1000 mg ÷ 500 mg/tab)

---

## Case 2: Liquid/Cream with Known Concentration (Continuous Dispense)
**Strength**: `strength_num` + `strength_unit_num` / `strength_den` + `strength_unit_den`  
**Dispense Unit**: `mL` or `g` (continuous)  
**Piece Content**: NULL

**Example**: Ibuprofen 100 mg/5 mL syrup
- Strength: `100 mg / 5 mL`
- Dispense Unit: `mL`

### Acceptable Prescription Units (dose_unit):
1. **`dispense_unit`** (e.g., `mL`, `g`) - Direct continuous dispensing
2. **`strength_unit_num`** (e.g., `mg`) - Convert to dispense_unit
3. **`strength_unit_den`** (e.g., `mL`) - Convert to dispense_unit

**Example Prescriptions**:
- `dose_unit = 'mL'`, `dose_amount = 10` → Dispense 10 mL
- `dose_unit = 'mg'`, `dose_amount = 200` → Dispense 10 mL (200 mg ÷ 100 mg/5 mL × 5 mL)
- `dose_unit = 'mL'`, `dose_amount = 5` → Dispense 5 mL (if dispense_unit = mL)

---

## Case 3: Liquid/Cream with Known Concentration (Bottle Dispense)
**Strength**: `strength_num` + `strength_unit_num` / `strength_den` + `strength_unit_den`  
**Dispense Unit**: `bottle`  
**Piece Content**: `piece_content_amount` + `piece_content_unit` (required, must match one of strength units)

**Example**: Amoxicillin 250 mg/5 mL suspension, 100 mL bottle
- Strength: `250 mg / 5 mL`
- Dispense Unit: `bottle`
- Piece Content: `100 mL`

### Acceptable Prescription Units (dose_unit):
1. **`bottle`** - Direct bottle dispensing
2. **`piece_content_unit`** (e.g., `mL`) - Convert to bottles
3. **`strength_unit_num`** (e.g., `mg`) - Convert via concentration → piece_content → bottles
4. **`strength_unit_den`** (e.g., `mL`) - Convert via piece_content → bottles

**Example Prescriptions**:
- `dose_unit = 'bottle'`, `dose_amount = 1` → Dispense 1 bottle
- `dose_unit = 'mL'`, `dose_amount = 150` → Dispense 2 bottles (150 mL ÷ 100 mL/bottle)
- `dose_unit = 'mg'`, `dose_amount = 500` → Convert: 500 mg → 10 mL → 0.1 bottle → 1 bottle (CEIL)

---

## Case 4: Solid with Unknown Strength
**Strength**: ALL NULL  
**Dispense Unit**: `tab`, `cap`, or `drop`  
**Piece Content**: NULL

**Example**: Unknown strength tablet
- Strength: NULL
- Dispense Unit: `tab`

### Acceptable Prescription Units (dose_unit):
1. **`dispense_unit`** (e.g., `tab`, `cap`, `drop`) - Must match exactly

**Example Prescriptions**:
- `dose_unit = 'tab'`, `dose_amount = 2` → Dispense 2 tablets
- `dose_unit = 'mg'` → ❌ ERROR: Cannot convert without strength info

---

## Case 5: Liquid with Unknown Concentration (Bottle Dispense)
**Strength**: ALL NULL  
**Dispense Unit**: `bottle`  
**Piece Content**: NULL

**Example**: Unknown concentration liquid in bottles
- Strength: NULL
- Dispense Unit: `bottle`

### Acceptable Prescription Units (dose_unit):
1. **`bottle`** - Must match exactly

**Example Prescriptions**:
- `dose_unit = 'bottle'`, `dose_amount = 2` → Dispense 2 bottles
- `dose_unit = 'mL'` → ❌ ERROR: Cannot convert without concentration info

---

## Summary Table

| Case | Strength | Dispense Unit | Piece Content | Acceptable dose_unit |
|------|----------|---------------|---------------|---------------------|
| 1. Solid (known) | `X mg/tab` | `tab`/`cap`/`drop` | NULL | `dispense_unit`, `strength_unit_num` |
| 2. Liquid (known, continuous) | `X mg/Y mL` | `mL`/`g` | NULL | `dispense_unit`, `strength_unit_num`, `strength_unit_den` |
| 3. Liquid (known, bottle) | `X mg/Y mL` | `bottle` | Required | `bottle`, `piece_content_unit`, `strength_unit_num`, `strength_unit_den` |
| 4. Solid (unknown) | NULL | `tab`/`cap`/`drop` | NULL | `dispense_unit` only |
| 5. Liquid (unknown, bottle) | NULL | `bottle` | NULL | `bottle` only |

---

## Key Rules

1. **Unknown strength** = No conversions possible, must match `dispense_unit` exactly
2. **Known strength solids** = Can convert from `strength_unit_num` to pieces
3. **Known concentration liquids (continuous)** = Can convert between `strength_unit_num` ↔ `strength_unit_den` ↔ `dispense_unit`
4. **Known concentration liquids (bottle)** = Can convert via concentration → `piece_content` → bottles
5. **Piece content** is only required for known-concentration liquids dispensed as bottles

