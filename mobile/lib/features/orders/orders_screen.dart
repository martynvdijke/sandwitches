import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers.dart';

class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref){
    final ordersAsync = ref.watch(ordersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Orders')),
      body: ordersAsync.when(
        data: (list){
          if (list.isEmpty) return const Center(child: Text('No orders'));
          return ListView.builder(itemCount: list.length, itemBuilder: (c,i){
            final o = list[i];
            return ListTile(
              title: Text('Order #${o.id} - ${o.status}'),
              subtitle: Text('\$${o.totalPrice}  ${o.createdAt}'),
              onTap: ()=> context.push('/orders/track/${o.trackingToken.isNotEmpty? o.trackingToken : o.id}'),
            );
          });
        },
        loading: ()=> const Center(child: CircularProgressIndicator()),
        error: (e,s)=> Center(child: Text('Error: $e')),
      ),
    );
  }
}
