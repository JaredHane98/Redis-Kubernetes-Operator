// import necessary module
import { check } from 'k6';
import { SharedArray } from 'k6/data';
import http from 'k6/http';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';



const data = new SharedArray('employees', function () {
  const f = JSON.parse(open('./output-database.json')).employees;
  return f; 
});


export const options = {
    thresholds: {
      http_req_failed: [{threshold:'rate<0.01', abortOnFail: false}],
      http_req_duration: ['p(99)<1000'], 
    },
    scenarios: {
      average_load: {
        executor: 'ramping-vus',
        startVUs: 1,
        stages: [
          { duration: '15m', target: '1000' },
        ],
      },
    },
  };

let index = 0


const env_url = __ENV.TARGET_URL;
if (!env_url) {
  throw new Error('TARGET_URL environment variable is not set.');
}
const url = "http://" + env_url + "/employee"

export default function () {

  let employee = Object.assign({}, data[index]);
  employee.id = uuidv4();
  index = (index + 1) % data.length;


  const payload = JSON.stringify(employee);

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  // send a post request and save response as a variable
  const res = http.post(url, payload, params);
  
  check(res, {
    'response code was 200': (res) => res.status == 200, 
  })
}
