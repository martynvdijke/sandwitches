/// Returns true when the current user may use admin-only features.
///
/// The backend exposes these flags on /api/v1/me. Keep this in one place so the
/// home action, the admin queue gate and the recipe form cannot drift apart.
bool isAdmin(Map<String, dynamic>? me) {
  if (me == null) return false;
  return me['is_staff'] == true || me['is_superuser'] == true;
}
