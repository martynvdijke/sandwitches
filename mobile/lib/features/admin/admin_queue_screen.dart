import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../../core/roles.dart';

class AdminQueueScreen extends ConsumerStatefulWidget {
  const AdminQueueScreen({super.key});
  @override
  ConsumerState<AdminQueueScreen> createState() => _AdminState();
}

class _AdminState extends ConsumerState<AdminQueueScreen> {
  bool? isStaff;
  @override
  void initState() {
    super.initState();
    _checkStaff();
  }

  Future<void> _checkStaff() async {
    final me = await ref.read(repositoryProvider).fetchMe();
    if (me != null) {
      setState(() => isStaff = isAdmin(me));
    } else {
      setState(() => isStaff = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (isStaff == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Admin Queue')),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    if (isStaff == false) {
      return Scaffold(
        appBar: AppBar(title: const Text('Admin Queue')),
        body: const Center(
          child: Text(
            'Admins only — you do not have permission to view the chef queue.',
          ),
        ),
      );
    }
    final ordersAsync = ref.watch(ordersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Chef Queue')),
      body: ordersAsync.when(
        data: (list) {
          if (list.isEmpty) return const Center(child: Text('No orders'));
          return ListView.builder(
            itemCount: list.length,
            itemBuilder: (c, i) {
              final o = list[i];
              return Card(
                margin: const EdgeInsets.all(8),
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Order #${o.id} - ${o.status}',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      Text('Total: \$${o.totalPrice}'),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        children: [
                          for (final s in [
                            'PENDING',
                            'MAKING',
                            'READY',
                            'DELIVERED',
                          ])
                            ElevatedButton(
                              onPressed: o.status.toUpperCase() == s
                                  ? null
                                  : () async {
                                      try {
                                        await ref
                                            .read(repositoryProvider)
                                            .patchOrderStatus(o.id, s);
                                        ref.invalidate(ordersProvider);
                                        if (context.mounted) {
                                          ScaffoldMessenger.of(context)
                                              .showSnackBar(
                                                SnackBar(
                                                  content: Text(
                                                    'Status updated to $s',
                                                  ),
                                                ),
                                              );
                                        }
                                      } catch (e) {
                                        if (context.mounted) {
                                          ScaffoldMessenger.of(context)
                                              .showSnackBar(
                                                SnackBar(
                                                  content: Text('Failed: $e'),
                                                ),
                                              );
                                        }
                                      }
                                    },
                              child: Text(s),
                            ),
                        ],
                      ),
                    ],
                  ),
                ),
              );
            },
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, s) => Center(child: Text('Error: $e')),
      ),
    );
  }
}
