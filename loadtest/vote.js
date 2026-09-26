import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

export const options = {
    vus: 100,
    duration: '30s',
};

const BASE_URL = 'http://localhost:8080';
const POLL_ID = 3;
const OPTION_ID = 7;

const status201 = new Counter('status_201');
const status400 = new Counter('status_400');
const status409 = new Counter('status_409');
const status429 = new Counter('status_429');
const status500 = new Counter('status_500');
const statusOther = new Counter('status_other');

export default function () {
    const clientID = `loadtest-${__VU}-${__ITER}`;

    const payload = JSON.stringify({
        option_id: OPTION_ID,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Cookie': `client_id=${clientID}`,
        },
    };

    const response = http.post(
        `${BASE_URL}/api/v1/polls/${POLL_ID}/vote`,
        payload,
        params,
    );

    switch (response.status) {
        case 201:
            status201.add(1);
            break;
        case 400:
            status400.add(1);
            break;
        case 409:
            status409.add(1);
            break;
        case 429:
            status429.add(1);
            break;
        case 500:
            status500.add(1);
            break;
        default:
            statusOther.add(1);
    }

    check(response, {
        'status is 201 or 409': (res) =>
            res.status === 201 || res.status === 409,
    });
}