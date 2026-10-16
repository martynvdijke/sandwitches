import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';

class AdminQueueScreen extends ConsumerStatefulWidget {
  const AdminQueueScreen({super.key});
  @override
  ConsumerState<AdminQueueScreen> createState()=> _AdminState();
}
class _AdminState extends ConsumerState<AdminQueueScreen> {
  bool? isStaff;
  @override
  void initState(){
    super.initState();
    _checkStaff();
  }
  Future<void> _checkStaff() async {
    // Detect staff by attempting to see if PATCH would be allowed? For now check /me for is_staff or is_superuser or role
    final me = await ref.read(repositoryProvider).fetchMe();
    if (me!=null) {
      final staff = me['is_staff']==true || me['is_superuser']==true || me['role']?.toString().toLowerCase()=='staff' || me['is_admin']==true;
      // Also attempt GET /orders and see if we get admin view? We'll just use me check fallback to allow if me exists
      // If no explicit flag, we still show orders but gate PATCH buttons
      setState(()=> isStaff = staff);
      if (!staff) {
        // Keep isStaff false to show admins only state, but still allow detection via permission check:
        // Try a dummy check: we consider user non-staff if flag missing
      }
    } else {
      setState(()=> isStaff=false);
    }
  }

  @override
  Widget build(BuildContext context){
    if (isStaff==null) return Scaffold(appBar: AppBar(title: const Text('Admin Queue')), body: const Center(child: CircularProgressIndicator()));
    if (isStaff==false) {
      return Scaffold(appBar: AppBar(title: const Text('Admin Queue')), body: const Center(child: Text('Admins only — you do not have permission to view the chef queue.')));
    }
    final ordersAsync = ref.watch(ordersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Chef Queue')),
      body: ordersAsync.when(
        data: (list){
          if (list.isEmpty) return const Center(child: Text('No orders'));
          return ListView.builder(itemCount: list.length, itemBuilder: (c,i){
            final o = list[i];
            return Card(margin: const EdgeInsets.all(8), child: Padding(padding: const EdgeInsets.all(12), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children:[
              Text('Order #${o.id} - ${o.status}', style: Theme.of(context).textTheme.titleMedium),
              Text('Total: \$${o.totalPrice}'),
              const SizedBox(height:8),
              Wrap(spacing:8, children:[
                for (final s in ['PENDING','MAKING','READY','DELIVERED']) ElevatedButton(
                  onPressed: o.status.toUpperCase()==s? null: () async {
                    try { await ref.read(repositoryProvider).patchOrderStatus(o.id, s); ref.invalidate(ordersProvider); if (context.mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Status updated to $s'))); } catch(e){ if(context.mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e'))); }
                  },
                  child: Text(s),
                ),
              ]),
            ])));
          });
        },
        loading: ()=> const Center(child: CircularProgressIndicator()),
        error: (e,s)=> Center(child: Text('Error: $e')),
      ),
    );
  }
}
