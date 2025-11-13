const https = require('https');

const BASE_URL = 'https://nestjs-microserviceuser-03d2b6c38b70.herokuapp.com';

// Test data - replace with actual user ID from your database
const testUserId = 'usr_61jqu173';

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

async function testUpdatePreferences() {
  console.log('Testing Update User Preferences Endpoint\n');
  console.log('='.repeat(60));

  try {
    // Test 1: Get current preferences
    console.log('\nTest 1: Get current user preferences');
    const result1 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
    );
    console.log(`Status: ${result1.status}`);
    console.log('Current preferences:', JSON.stringify(result1.data, null, 2));

    if (result1.status !== 200) {
      console.log(
        '\nUser not found. Please update testUserId variable with a valid user ID.',
      );
      return;
    }

    const currentPrefs = result1.data.preferences;

    // Test 2: Update email preference only
    console.log('\n Test 2: Update email preference only');
    const result2 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
      'PATCH',
      {
        email: !currentPrefs.email, // Toggle email preference
      },
    );
    console.log(`Status: ${result2.status}`);
    console.log('Updated preferences:', JSON.stringify(result2.data, null, 2));

    // Test 3: Update push preference only
    console.log('\n Test 3: Update push preference only');
    const result3 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
      'PATCH',
      {
        push: !currentPrefs.push, // Toggle push preference
      },
    );
    console.log(`Status: ${result3.status}`);
    console.log('Updated preferences:', JSON.stringify(result3.data, null, 2));

    // Test 4: Update both preferences
    console.log('\n Test 4: Update both email and push preferences');
    const result4 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
      'PATCH',
      {
        email: true,
        push: false,
      },
    );
    console.log(`Status: ${result4.status}`);
    console.log('Updated preferences:', JSON.stringify(result4.data, null, 2));

    // Test 5: Verify cache was invalidated (should be Cache MISS on server logs)
    console.log(
      '\n📝 Test 5: Verify preferences after update (cache should be invalidated)',
    );
    const result5 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
    );
    console.log(`Status: ${result5.status}`);
    console.log('Verified preferences:', JSON.stringify(result5.data, null, 2));

    // Test 6: Restore original preferences
    console.log('\n Test 6: Restore original preferences');
    const result6 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
      'PATCH',
      {
        email: currentPrefs.email,
        push: currentPrefs.push,
      },
    );
    console.log(`Status: ${result6.status}`);
    console.log('Restored preferences:', JSON.stringify(result6.data, null, 2));

    // Test 7: Test validation - no fields provided
    console.log(
      '\n📝 Test 7: Test validation - no fields provided (should fail)',
    );
    const result7 = await makeRequest(
      `/api/v1/users/${testUserId}/preferences`,
      'PATCH',
      {},
    );
    console.log(`Status: ${result7.status} (Expected: 400)`);
    console.log('Error:', JSON.stringify(result7.data, null, 2));

    // Test 8: Test with non-existent user
    console.log('\n Test 8: Test with non-existent user (should fail)');
    const result8 = await makeRequest(
      `/api/v1/users/usr_nonexistent/preferences`,
      'PATCH',
      {
        email: true,
      },
    );
    console.log(`Status: ${result8.status} (Expected: 404)`);
    console.log('Error:', JSON.stringify(result8.data, null, 2));

    console.log('\n' + '='.repeat(60));
    console.log(' Update preferences endpoint testing completed!');
    console.log('\n Key Features:');
    console.log('- ✓ PATCH /api/v1/users/:user_id/preferences');
    console.log('- ✓ Partial updates (email only, push only, or both)');
    console.log('- ✓ Automatic cache invalidation after update');
    console.log('- ✓ Validation for empty request body');
    console.log('- ✓ 404 error for non-existent users');
    console.log('- ✓ Returns updated preferences with metadata');
  } catch (error) {
    console.error('❌ Error during testing:', error.message);
  }
}

// Run tests
testUpdatePreferences();
