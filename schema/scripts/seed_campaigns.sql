-- Seed campaigns

insert into campaigns (name, channel, status, base_template, scheduled_at)
values (
  'Welcome Offer',
  'sms',
  'draft',
  'Hi {FirstName}, welcome to our store! Use code WELCOME10 for 10% off.',
  null
);

insert into campaigns (name, channel, status, base_template, scheduled_at)
values (
  'Festive Sale',
  'whatsapp',
  'scheduled',
  'Hello {FirstName}, our Festive Sale is live — get 25% off {PreferredProduct} at our {Location} store!',
  '2025-12-15 10:00:00'
);

insert into campaigns (name, channel, status, base_template, scheduled_at)
values (
  'Restock Alert',
  'sms',
  'draft',
  'Hi {FirstName}, {PreferredProduct} is back in stock at {Location}. Grab yours while stocks last!',
  null
);
