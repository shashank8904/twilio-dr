async function run() {
  const agentUrl = 'http://localhost:8080/dr/agents';
  const enqueueUrl = 'http://localhost:8080/dr/enqueue';
  const queueUrl = 'http://localhost:8080/dr/queue';
  const acceptUrl = 'http://localhost:8080/dr/accept';
  const completeUrl = 'http://localhost:8080/dr/call-status';
  const mockUrl = 'http://localhost:8080/mock/telephony/conferences';

  console.log('1. Creating agent...');
  const agentRes = await fetch(agentUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: 'test-agent-' + Date.now(),
      displayName: 'Test Agent',
      languages: ['EN'],
      domains: ['Support'],
      status: 'available',
      maxConcurrentCalls: 1
    })
  });
  const agent = await agentRes.json();
  console.log('Agent:', agent);

  console.log('2. Enqueuing call...');
  const enqueueRes = await fetch(enqueueUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      callSid: 'CA-TEST-' + Date.now(),
      caller: '+1234567890',
      called: '+0987654321',
      language: 'EN',
      domain: 'Support'
    })
  });
  const call = await enqueueRes.json();
  console.log('Call:', call);

  console.log('3. Checking queue...');
  const queueRes = await fetch(`${queueUrl}?agentId=${agent.id}`);
  const queue = await queueRes.json();
  console.log('Queue:', queue);

  console.log('4. Accepting call...');
  const acceptRes = await fetch(acceptUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      callSid: call.callSid,
      agentId: agent.id
    })
  });
  const accepted = await acceptRes.json();
  console.log('Accepted:', accepted);

  console.log('5. Checking mock telephony...');
  const mockRes = await fetch(mockUrl);
  const mock = await mockRes.json();
  console.log('Mock Telephony:', JSON.stringify(mock, null, 2));

  console.log('6. Completing call...');
  const completeRes = await fetch(completeUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      callSid: call.callSid,
      status: 'completed'
    })
  });
  const completed = await completeRes.json();
  console.log('Completed:', completed);
}
run().catch(console.error);
