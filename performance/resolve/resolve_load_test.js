import http from 'k6/http';
import { check } from 'k6';

export const options = {
  maxRedirects: 0,
  scenarios: {
    resolve_load: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 1000,
      maxVUs: 10000,
    },
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default function () {
  const res = http.get('http://haproxy/1J9Mt4');
  check(res, {
    'status is 302': (r) => r.status === 302,
  });
}