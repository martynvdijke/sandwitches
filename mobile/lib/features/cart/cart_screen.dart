import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers.dart';

class CartScreen extends ConsumerWidget {
  const CartScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref){
    final cartAsync = ref.watch(cartProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Cart')),
      body: cartAsync.when(
        data: (items){
          if (items.isEmpty) return const Center(child: Text('Cart is empty'));
          double total=0;
          for (final it in items) { total += (double.tryParse(it.recipe.price) ?? 0) * it.quantity; }
          return Column(children:[
            Expanded(child: ListView.builder(itemCount: items.length, itemBuilder: (c,i){
              final it = items[i];
              return ListTile(
                title: Text(it.recipe.title),
                subtitle: Text('\$${it.recipe.price} x ${it.quantity}'),
                trailing: Row(mainAxisSize: MainAxisSize.min, children:[
                  IconButton(icon: const Icon(Icons.remove), onPressed: () async {
                    final nq = it.quantity-1;
                    if (nq<=0) { await ref.read(repositoryProvider).deleteCart(it.id); } else { await ref.read(repositoryProvider).updateCart(it.id, nq); }
                    ref.invalidate(cartProvider);
                  }),
                  Text('${it.quantity}'),
                  IconButton(icon: const Icon(Icons.add), onPressed: () async {
                    await ref.read(repositoryProvider).updateCart(it.id, it.quantity+1);
                    ref.invalidate(cartProvider);
                  }),
                  IconButton(icon: const Icon(Icons.delete), onPressed: () async {
                    await ref.read(repositoryProvider).deleteCart(it.id);
                    ref.invalidate(cartProvider);
                  }),
                ]),
              );
            })),
            Padding(padding: const EdgeInsets.all(16), child: Column(children:[
              Text('Total: \$${total.toStringAsFixed(2)}', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height:8),
              ElevatedButton(onPressed: () async {
                if (items.isEmpty) return;
                try {
                  // Checkout: create order for first item (spec says POST /orders body recipe_id)
                  final o = await ref.read(repositoryProvider).createOrder(items.first.recipe.id);
                  ref.invalidate(cartProvider);
                  if (context.mounted) context.push('/orders/track/${o.trackingToken.isNotEmpty? o.trackingToken : o.id}');
                } catch(e){
                  if (context.mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Checkout failed: $e')));
                }
              }, child: const Text('Checkout')),
            ])),
          ]);
        },
        loading: ()=> const Center(child: CircularProgressIndicator()),
        error: (e,s)=> Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children:[Text('Error: $e'), ElevatedButton(onPressed: ()=> ref.invalidate(cartProvider), child: const Text('Retry'))])),
      ),
    );
  }
}
