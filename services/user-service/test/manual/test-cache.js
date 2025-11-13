const https = require('https');

const BASE_URL = 'https://nestjs-microserviceuser-03d2b6c38b70.herokuapp.com';

// Test data
const testUserId = 'usr_test123'; // Replace with actual user ID from your database

// Helper function to make HTTP requests
function makeRequest(path, method = 'GET', data = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, BASE_URL);
    const options = {
      method: method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    const req = https.request(url, options, (res) => {
      let body = '';
      res.on('data', (chunk) => (body += chunk));
      res.on('end', () => {
        try {
          resolve({
            status: res.statusCode,
            data: JSON.parse(body),
          });
        } catch (e) {
          resolve({
            status: res.statusCode,
            data: body,
          });
        }
      });
    });

    req.on('error', reject);
    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

async function testCaching() {
  console.log('🧪 Testing Redis Cache Implementation\n');
  console.log('='.repeat(60));

  try {
    // Test 1: Get user preferences (Cache MISS)
    console.log('\n📝 Test 1: First request (Cache MISS expected)');
    const start1 = Date.now();
    const result1 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
    );
    const time1 = Date.now() - start1;
    console.log(`Status: ${result1.status}`);
    console.log(`Response time: ${time1}ms`);
    console.log('Data:', JSON.stringify(result1.data, null, 2));

    // Test 2: Get same user preferences (Cache HIT)
    console.log('\n📝 Test 2: Second request (Cache HIT expected)');
    const start2 = Date.now();
    const result2 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
    );
    const time2 = Date.now() - start2;
    console.log(`Status: ${result2.status}`);
    console.log(`Response time: ${time2}ms`);
    console.log('Data:', JSON.stringify(result2.data, null, 2));

    console.log(
      `\n⚡ Speed improvement: ${(((time1 - time2) / time1) * 100).toFixed(1)}%`,
    );

    // Test 3: Batch get users
    console.log('\n📝 Test 3: Batch get users (partial cache HIT expected)');
    const start3 = Date.now();
    const result3 = await makeRequest(
      '/api/v1/users/preferences/batch',
      'POST',
      {
        user_ids: [testUserId, 'usr_another', 'usr_test456'],
      },
    );
    const time3 = Date.now() - start3;
    console.log(`Status: ${result3.status}`);
    console.log(`Response time: ${time3}ms`);
    console.log('Data:', JSON.stringify(result3.data, null, 2));

    // Test 4: Update notification time (Cache invalidation)
    console.log('\n📝 Test 4: Update notification time (Cache invalidation)');
    const result4 = await makeRequest(
      `/api/v1/users/${testUserId}/last-notification`,
      'POST',
      {
        channel: 'email',
        sent_at: new Date().toISOString(),
        notification_id: 'notif_' + Date.now(),
      },
    );
    console.log(`Status: ${result4.status}`);
    console.log('Cache should be invalidated now');

    // Test 5: Get user preferences again (Cache MISS after invalidation)
    console.log(
      '\n📝 Test 5: Request after cache invalidation (Cache MISS expected)',
    );
    const start5 = Date.now();
    const result5 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
    );
    const time5 = Date.now() - start5;
    console.log(`Status: ${result5.status}`);
    console.log(`Response time: ${time5}ms`);
    console.log('Data:', JSON.stringify(result5.data, null, 2));

    console.log('\n' + '='.repeat(60));
    console.log('✅ Cache testing completed!');
    console.log('\n💡 Tips:');
    console.log(
      '- Check server logs for "Cache HIT" and "Cache MISS" messages',
    );
    console.log('- Cache HIT requests should be significantly faster');
    console.log('- Cache is automatically invalidated on updates');
    console.log('- Set REDIS_URL env var to use Redis in production');
    console.log('- Without REDIS_URL, in-memory cache is used (dev mode)');
  } catch (error) {
    console.error('❌ Error during testing:', error.message);
  }
}

// Run tests
testCaching();
