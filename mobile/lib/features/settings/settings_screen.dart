import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../core/config.dart';
import '../../core/providers.dart';

/// Settings and legal actions. The privacy policy and terms are served by the
/// configured Sandwitches instance, so a self-hosted deployment publishes its
/// own controller details rather than the default instance's.
class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  Future<void> _openLegal(
    BuildContext context,
    WidgetRef ref,
    String path,
  ) async {
    final saved = await ref.read(storageProvider).getBaseUrl();
    final raw = ref.read(baseUrlProvider) ?? saved ?? kDefaultBaseUrl;
    final base = raw.endsWith('/') ? raw.substring(0, raw.length - 1) : raw;
    final uri = Uri.tryParse('$base$path');
    if (uri == null ||
        !await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Could not open the page in a browser')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
            onTap: () => _openLegal(context, ref, '/privacy'),
          ),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: const Text('Terms of use'),
            subtitle: const Text('Rules for using this instance'),
            onTap: () => _openLegal(context, ref, '/terms'),
          ),
          ListTile(
            leading: const Icon(Icons.info_outline),
            title: const Text('Delete account information'),
            subtitle: const Text('What is removed and what is retained'),
            onTap: () => _openLegal(context, ref, '/account/delete'),
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text('Account'),
          ),
          ListTile(
            leading: const Icon(Icons.download_outlined),
            title: const Text('Download my data'),
            subtitle: const Text('JSON export of your account and activity'),
            onTap: () => _openLegal(context, ref, '/api/v1/me/export'),
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
                      style: TextStyle(color: Theme.of(context).colorScheme.error),
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
