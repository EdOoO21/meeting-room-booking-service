import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: Number(__ENV.VUS || 50),
  duration: __ENV.DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200'],
  },
};

export function setup() {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8081';
  const targetDate = __ENV.TARGET_DATE || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 10);

  const loginRes = http.post(
    `${baseURL}/dummyLogin`,
    JSON.stringify({ role: 'user' }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(loginRes, { 'dummyLogin returns 200': (r) => r.status === 200 });

  const token = loginRes.json('token');
  const roomsRes = http.get(`${baseURL}/rooms/list`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  check(roomsRes, { 'rooms list returns 200': (r) => r.status === 200 });

  const rooms = roomsRes.json('rooms') || [];
  if (!rooms.length) {
    throw new Error('No rooms found. Run perf seed before load test.');
  }

  return {
    baseURL,
    token,
    roomId: rooms[0].id,
    targetDate,
  };
}

export default function (data) {
  const res = http.get(`${data.baseURL}/rooms/${data.roomId}/slots/list?date=${data.targetDate}`, {
    headers: {
      Authorization: `Bearer ${data.token}`,
    },
  });

  check(res, {
    'slots list returns 200': (r) => r.status === 200,
    'slots payload has array': (r) => Array.isArray(r.json('slots')),
  });

  sleep(1);
}
