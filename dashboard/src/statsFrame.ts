// The SDK's ModuleStats heartbeat shares the stats leaf and also puts a
// length-delimited message in field 1, so only a tag walk separates the two.

type Varint = { value: number; next: number };

function readVarint(b: Uint8Array, pos: number): Varint | null {
  let value = 0;
  let scale = 1;
  for (let i = pos; i < b.length; i++) {
    const byte = b[i];
    value += (byte & 0x7f) * scale;
    if ((byte & 0x80) === 0) return { value, next: i + 1 };
    if (i - pos >= 9) return null;
    scale *= 128;
  }
  return null;
}

/**
 * Reports whether a frame published on the stats leaf is an `UmrStats`.
 * Accepts only a buffer whose top-level tags include a length-delimited field
 * numbered 2 to 6; a `ModuleStats` heartbeat can carry field 1 (length-delimited)
 * and field 2 (varint) only, so it is always rejected. Malformed input is rejected.
 */
export function isUmrStatsFrame(bytes: Uint8Array): boolean {
  let pos = 0;
  while (pos < bytes.length) {
    const tag = readVarint(bytes, pos);
    if (tag === null) return false;
    pos = tag.next;
    const fieldNo = Math.floor(tag.value / 8);
    const wireType = tag.value % 8;
    if (fieldNo >= 2 && fieldNo <= 6 && wireType === 2) return true;

    switch (wireType) {
      case 0: {
        const v = readVarint(bytes, pos);
        if (v === null) return false;
        pos = v.next;
        break;
      }
      case 1:
        pos += 8;
        break;
      case 2: {
        const len = readVarint(bytes, pos);
        if (len === null) return false;
        pos = len.next + len.value;
        break;
      }
      case 5:
        pos += 4;
        break;
      default:
        return false;
    }
    if (pos > bytes.length) return false;
  }
  return false;
}
