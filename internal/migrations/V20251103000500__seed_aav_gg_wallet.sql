-- One-time credit for @aav_gg (username without @)
UPDATE wallets w
SET
    balance = balance + 1000,
    balance_available = balance_available + 1000,
    total_earned = total_earned + 1000
FROM profiles p
WHERE w.profile_id = p.id
  AND LOWER(TRIM(p.username)) = 'aav_gg';
