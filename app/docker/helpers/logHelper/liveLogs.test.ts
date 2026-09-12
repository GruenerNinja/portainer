import { DockerLogStreamDecoder } from './liveLogs';

function dockerFrame(text: string, stream = 1) {
  const payload = new TextEncoder().encode(text);
  const frame = new Uint8Array(8 + payload.length);
  frame[0] = stream;
  new DataView(frame.buffer).setUint32(4, payload.length);
  frame.set(payload, 8);
  return frame;
}

describe('DockerLogStreamDecoder', () => {
  it('decodes a raw TTY stream across UTF-8 chunks', () => {
    const decoder = new DockerLogStreamDecoder(false);
    const encoded = new TextEncoder().encode('ready ✓\n');

    expect(decoder.push(encoded.slice(0, encoded.length - 2))).toBe('ready ');
    expect(decoder.push(encoded.slice(encoded.length - 2))).toBe('✓\n');
    expect(decoder.flush()).toBe('');
  });

  it('demultiplexes Docker frames split across websocket messages', () => {
    const decoder = new DockerLogStreamDecoder(true);
    const first = dockerFrame('stdout\n');
    const second = dockerFrame('stderr\n', 2);
    const combined = new Uint8Array(first.length + second.length);
    combined.set(first);
    combined.set(second, first.length);

    expect(decoder.push(combined.slice(0, 5))).toBe('');
    expect(decoder.push(combined.slice(5, first.length + 3))).toBe('stdout\n');
    expect(decoder.push(combined.slice(first.length + 3))).toBe('stderr\n');
    expect(decoder.flush()).toBe('');
  });

  it('rejects oversized frames before buffering their payload', () => {
    const decoder = new DockerLogStreamDecoder(true);
    const header = new Uint8Array(8);
    new DataView(header.buffer).setUint32(4, 16 * 1024 * 1024 + 1);

    expect(() => decoder.push(header)).toThrow('safety limit');
  });

  it('rejects an incomplete multiplexed frame at end of stream', () => {
    const decoder = new DockerLogStreamDecoder(true);
    decoder.push(dockerFrame('partial').slice(0, 10));

    expect(() => decoder.flush()).toThrow('incomplete frame');
  });
});
