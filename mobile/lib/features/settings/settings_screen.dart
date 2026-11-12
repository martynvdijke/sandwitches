import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../core/config.dart';
import '../../core/models.dart';
import '../../core/notifications.dart';
import '../../core/providers.dart';

/// Settings and legal actions. The privacy policy and terms are served by the
/// configured Sandwitches instance, so a self-hosted deployment publishes its
/// own controller details rather than the default instance's.
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  bool _exporting = false;
  NotificationPreferences? _prefs;
  bool _prefsLoading = true;
  bool _prefsSaving = false;
  String? _prefsError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadPrefs());
  }

  Future<void> _loadPrefs() async {
    setState(() {
      _prefsLoading = true;
      _prefsError = null;
    });
    try {
      final prefs = await ref
          .read(repositoryProvider)
          .fetchNotificationPreferences();
      if (!mounted) return;
      setState(() {
        _prefs = prefs;
        _prefsLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _prefsLoading = false;
        _prefsError = e is ApiException
            ? e.message
            : 'Could not load notification settings';
      });
    }
  }

  /// Saves one toggle optimistically and rolls back when the server rejects it.
  Future<void> _setPref(
    NotificationPreferences next,
    Map<String, bool> change, {
    void Function()? afterSave,
  }) async {
    final previous = _prefs;
    setState(() {
      _prefs = next;
      _prefsSaving = true;
      _prefsError = null;
    });
    try {
      final saved = await ref
          .read(repositoryProvider)
          .updateNotificationPreferences(change);
      if (!mounted) return;
      setState(() => _prefs = saved);
      afterSave?.call();
    } catch (e) {
      if (!mounted) return;
      setState(() => _prefs = previous);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            e is ApiException ? e.message : 'Could not save notification settings',
          ),
        ),
      );
    } finally {
      if (mounted) setState(() => _prefsSaving = false);
    }
  }

  List<Widget> _notificationTiles() {
    if (_prefsLoading) {
      return const [
        ListTile(
          leading: SizedBox(
            width: 24,
            height: 24,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          title: Text('Loading notification settings'),
        ),
      ];
    }
    final prefs = _prefs;
    if (prefs == null) {
      return [
        ListTile(
          leading: const Icon(Icons.notifications_off_outlined),
          title: const Text('Notification settings unavailable'),
          subtitle: Text(_prefsError ?? 'Could not load notification settings'),
          trailing: TextButton(
            onPressed: _loadPrefs,
            child: const Text('Retry'),
          ),
        ),
      ];
    }
    return [
      SwitchListTile(
        secondary: const Icon(Icons.restaurant_menu),
        title: const Text('New recipes'),
        subtitle: const Text('When a recipe is added to this instance'),
        value: prefs.notifyNewRecipes,
        onChanged: _prefsSaving
            ? null
            : (value) => _setPref(
                prefs.copyWith(notifyNewRecipes: value),
                {'notify_new_recipes': value},
              ),
      ),
      SwitchListTile(
        secondary: const Icon(Icons.receipt_long_outlined),
        title: const Text('New orders'),
        subtitle: const Text('Order and status updates'),
        value: prefs.notifyNewOrders,
        onChanged: _prefsSaving
            ? null
            : (value) => _setPref(
                prefs.copyWith(notifyNewOrders: value),
                {'notify_new_orders': value},
              ),
      ),
      SwitchListTile(
        secondary: const Icon(Icons.language),
        title: const Text('In-app and web'),
        subtitle: const Text('Notifications while using Sandwitches'),
        value: prefs.notifyWeb,
        onChanged: _prefsSaving
            ? null
            : (value) => _setPref(
                prefs.copyWith(notifyWeb: value),
                {'notify_web': value},
              ),
      ),
      SwitchListTile(
        secondary: const Icon(Icons.notifications_active_outlined),
        title: const Text('Push notifications'),
        subtitle: const Text('Send notifications to this device'),
        value: prefs.notifyPush,
        onChanged: _prefsSaving
            ? null
            : (value) => _setPref(
                prefs.copyWith(notifyPush: value),
                {'notify_push': value},
                afterSave: value
                    ? () => NotificationService.instance.syncToken()
                    : null,
              ),
      ),
      if (_prefsError != null)
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 4, 16, 4),
          child: Text(
            _prefsError!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ),
    ];
  }

  Future<void> _openLegal(String path) async {
    final saved = await ref.read(storageProvider).getBaseUrl();
    final raw = ref.read(baseUrlProvider) ?? saved ?? kDefaultBaseUrl;
    final base = raw.endsWith('/') ? raw.substring(0, raw.length - 1) : raw;
    final uri = Uri.tryParse('$base$path');
    if (uri == null ||
        !await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Could not open the page in a browser')),
        );
      }
    }
  }

  Future<void> _exportData() async {
    if (_exporting) return;
    setState(() => _exporting = true);
    try {
      final bytes = await ref.read(repositoryProvider).exportAccount();
      final dir = await getApplicationDocumentsDirectory();
      final stamp = DateTime.now().toUtc().toIso8601String().replaceAll(
        RegExp(r'[:.]'),
        '-',
      );
      final file = File('${dir.path}/sandwitches-export-$stamp.json');
      await file.writeAsBytes(bytes, flush: true);
      await SharePlus.instance.share(
        ShareParams(
          files: [XFile(file.path, mimeType: 'application/json')],
          subject: 'Sandwitches data export',
        ),
      );
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('Data export ready')));
    } catch (e) {
      if (!mounted) return;
      final message = e is ApiException
          ? e.message
          : 'Could not export your data';
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
    } finally {
      if (mounted) setState(() => _exporting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final meAsync = ref.watch(meProvider);
    final username = meAsync.maybeWhen(
      data: (me) => me?['username']?.toString(),
      orElse: () => null,
    );
    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: ListView(
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text('Legal'),
          ),
          ListTile(
            leading: const Icon(Icons.privacy_tip_outlined),
            title: const Text('Privacy policy'),
            subtitle: const Text('What is stored, why, and who receives it'),
            onTap: () => _openLegal('/privacy'),
          ),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: const Text('Terms of use'),
            subtitle: const Text('Rules for using this instance'),
            onTap: () => _openLegal('/terms'),
          ),
          ListTile(
            leading: const Icon(Icons.info_outline),
            title: const Text('Delete account information'),
            subtitle: const Text('What is removed and what is retained'),
            onTap: () => _openLegal('/account/delete'),
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text('Account'),
          ),
          ListTile(
            leading: _exporting
                ? const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.download_outlined),
            title: const Text('Download my data'),
            subtitle: const Text('JSON export of your account and activity'),
            enabled: !_exporting,
            onTap: _exporting ? null : _exportData,
          ),
          ListTile(
            leading: Icon(
              Icons.delete_forever,
              color: Theme.of(context).colorScheme.error,
            ),
            title: Text(
              'Delete account',
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
            subtitle: const Text('Permanently close this account'),
            onTap: username == null
                ? null
                : () => _DeleteAccountDialog.show(context, ref, username),
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text('Notifications'),
          ),
          ..._notificationTiles(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }
}

class _DeleteAccountDialog extends StatefulWidget {
  final String username;
  const _DeleteAccountDialog({required this.username});

  static Future<void> show(
    BuildContext context,
    WidgetRef ref,
    String username,
  ) {
    return showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (_) => _DeleteAccountDialog(username: username),
    );
  }

  @override
  State<_DeleteAccountDialog> createState() => _DeleteAccountDialogState();
}

class _DeleteAccountDialogState extends State<_DeleteAccountDialog> {
  static const _effects = [
    'Your email, username, name, bio, avatar and password are erased.',
    'Every session is invalidated, including this phone.',
    'Your ratings, comments, favorites and cart are deleted.',
    'Recipes you authored stay public, shown without an author.',
    'Orders stay as anonymous transaction records.',
    'OIDC/Authelia accounts are not deleted at the provider.',
    'This cannot be undone.',
  ];

  final _confirm = TextEditingController();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _confirm.dispose();
    super.dispose();
  }

  Future<void> _delete(WidgetRef ref) async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(repositoryProvider).deleteAccount(widget.username);
      await NotificationService.instance.unregisterToken();
      await ref.read(storageProvider).clearAuth();
      await ref.read(persistCookieJarProvider).deleteAll();
      ref.invalidate(meProvider);
      if (!mounted) return;
      Navigator.of(context).pop();
      if (mounted) context.go('/login');
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer(
      builder: (context, ref, _) {
        final matches = _confirm.text.trim() == widget.username;
        return AlertDialog(
          title: const Text('Delete your account?'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                for (final effect in _effects)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('• '),
                        Expanded(child: Text(effect)),
                      ],
                    ),
                  ),
                const SizedBox(height: 12),
                Text('Type ${widget.username} to confirm.'),
                const SizedBox(height: 8),
                TextField(
                  controller: _confirm,
                  enabled: !_loading,
                  autofocus: true,
                  decoration: const InputDecoration(
                    labelText: 'Username',
                    border: OutlineInputBorder(),
                  ),
                  onChanged: (_) => setState(() {}),
                ),
                if (_error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      _error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: _loading ? null : () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error,
              ),
              onPressed: (!matches || _loading) ? null : () => _delete(ref),
              child: _loading
                  ? const SizedBox(
                      height: 18,
                      width: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Text('Delete account'),
            ),
          ],
        );
      },
    );
  }
}
