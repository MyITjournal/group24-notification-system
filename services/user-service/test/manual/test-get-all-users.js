// Test GET all users endpoint
const testGetAllUsers = async () => {
  console.log('=== Testing GET All Simple Users ===\n');

  try {
    // First, create a couple of test users
    console.log('1. Creating test users...');

    const user1 = await fetch('http://localhost:8000/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: 'Alice Smith',
        email: 'alice' + Date.now() + '@example.com',
        password: 'password123',
        preferences: { email: true, push: true },
      }),
    });
    const user1Data = await user1.json();
    console.log('✓ Created user 1:', user1Data.name, user1Data.email);

    const user2 = await fetch('http://localhost:8000/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: 'Bob Johnson',
        email: 'bob' + Date.now() + '@example.com',
        password: 'password123',
        preferences: { email: false, push: true },
      }),
    });
    const user2Data = await user2.json();
    console.log('✓ Created user 2:', user2Data.name, user2Data.email);

    // Now get all users
    console.log('\n2. Fetching all users...');
    const response = await fetch('http://localhost:8000/api/v1/users');

    if (response.ok) {
      const users = await response.json();
      console.log(`✓ Status: ${response.status}`);
      console.log(`✓ Total users: ${users.length}`);
      console.log('\nUsers:');
      users.forEach((user, index) => {
        console.log(`\n${index + 1}. ${user.email}`);
        console.log(`   - User ID: ${user.user_id}`);
        console.log(`   - Email Preference: ${user.preferences.email}`);
        console.log(`   - Push Preference: ${user.preferences.push}`);
        console.log(
          `   - Created: ${new Date(user.updated_at).toLocaleString()}`,
        );
        if (user.last_notification_email) {
          console.log(
            `   - Last Email: ${new Date(user.last_notification_email).toLocaleString()}`,
          );
        }
        if (user.last_notification_push) {
          console.log(
            `   - Last Push: ${new Date(user.last_notification_push).toLocaleString()}`,
          );
        }
      });

      console.log('\n=== Test Passed! ===');
    } else {
      const error = await response.text();
      console.error('✗ Failed:', response.status, error);
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
};

testGetAllUsers();
