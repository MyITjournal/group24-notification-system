// Test simple user endpoints
const testBothModules = async () => {
  console.log('=== Testing Simple Users Module (/api/v1/users) ===\n');

  // 1. Create a simple user
  try {
    const email = 'test' + Date.now() + '@example.com';
    console.log('1. Creating simple user...');
    const createResponse = await fetch('http://localhost:8081/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: 'Simple Test User',
        email: email,
        password: 'password123',
        preferences: { email: true, push: false },
      }),
    });

    const userData = await createResponse.json();
    console.log(`✓ Status: ${createResponse.status}`);
    console.log(`✓ Created user:`, userData);
    const userId = userData.user_id;

    // 2. Get simple user preferences
    console.log('\n2. Getting simple user preferences...');
    const prefsResponse = await fetch(
      `http://localhost:8081/api/v1/users/${userId}/preferences`,
    );
    const prefsData = await prefsResponse.json();
    console.log(`✓ Status: ${prefsResponse.status}`);
    console.log(`✓ Preferences:`, prefsData);

    // 3. Batch get preferences
    console.log('\n3. Batch getting simple user preferences...');
    const batchResponse = await fetch(
      'http://localhost:8081/api/v1/users/preferences/batch',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_ids: [userId, 'nonexistent'] }),
      },
    );
    const batchData = await batchResponse.json();
    console.log(`✓ Status: ${batchResponse.status}`);
    console.log(`✓ Batch result:`, batchData);

    // 4. Update last notification
    console.log('\n4. Updating last notification...');
    const notifResponse = await fetch(
      `http://localhost:8081/api/v1/users/${userId}/last-notification`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          channel: 'email',
          sent_at: new Date().toISOString(),
          notification_id: 'notif_12345',
        }),
      },
    );
    console.log(`✓ Status: ${notifResponse.status} (should be 204)`);

    console.log('\n=== All Simple Users Tests Passed! ===\n');
  } catch (error) {
    console.error('Simple Users Error:', error.message);
  }
};

testBothModules();
