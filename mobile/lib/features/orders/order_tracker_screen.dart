import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';
import '../../core/models.dart';

class OrderTrackerScreen extends ConsumerStatefulWidget {
  final String token;
  const OrderTrackerScreen({super.key, required this.token});
  @override
  ConsumerState<OrderTrackerScreen> createState()=> _TrackerState();
}
class _TrackerState extends ConsumerState<OrderTrackerScreen> {
  OrderDetail? order;
  String? error;
  bool loading=true;
  Timer? timer;
  static const statuses = ['PENDING','MAKING','READY','DELIVERED'];
  @override
  void initState(){
    super.initState();
    _fetch();
    timer = Timer.periodic(const Duration(seconds:5), (_)=> _fetch());
  }
  Future<void> _fetch() async {
    try {
      late OrderDetail d;
      bool fetched = false;
      try {
        d = await ref.read(repositoryProvider).trackOrder(widget.token);
        fetched = true;
      } catch (_) {
        final id = int.tryParse(widget.token);
        if (id!=null) {
          d = await ref.read(repositoryProvider).fetchOrder(id);
          fetched = true;
        }
      }
      if (!fetched) {
        throw Exception('Failed to fetch order');
      }
      if (mounted) {
        setState(() {
          order=d;
          loading=false;
          error=null;
        });
      }
    } catch(e){
      if (mounted) {
        setState(() {
          error=e.toString();
          loading=false;
        });
      }
    }
  }
  @override
  void dispose(){ timer?.cancel(); super.dispose(); }
  int _statusIndex(String s){
    final u = s.toUpperCase();
    final i = statuses.indexWhere((e)=> e==u);
    if (i==-1) {
      return 0;
    }
    return i;
  }
  @override
  Widget build(BuildContext context){
    if (loading) {
      return Scaffold(appBar: AppBar(title: const Text('Order Tracker')), body: const Center(child: CircularProgressIndicator()));
    }
    if (error!=null) {
      return Scaffold(appBar: AppBar(title: const Text('Order Tracker')), body: Center(child: Text('Error: $error')));
    }
    if (order==null) {
      return Scaffold(appBar: AppBar(title: const Text('Order Tracker')), body: const Center(child: Text('No order')));
    }
    return Scaffold(
      appBar: AppBar(title: const Text('Order Tracker')),
      body: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children:[
        Text('Order #${order!.id}', style: Theme.of(context).textTheme.titleLarge),
        Text('Status: ${order!.status}', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height:16),
        Row(children: List.generate(statuses.length, (i){
          final cur = _statusIndex(order!.status);
          final done = i <= cur;
          return Expanded(child: Column(children:[
            Icon(done? Icons.check_circle: Icons.circle_outlined, color: done? Colors.green: Colors.grey),
            Text(statuses[i], style: TextStyle(fontSize:10, fontWeight: done? FontWeight.bold: FontWeight.normal)),
            if (i<statuses.length-1) Container(height:4, color: i<cur? Colors.green: Colors.grey[300]),
          ]));
        })),
        const SizedBox(height:16),
        Text('Total: \$${order!.totalPrice}'),
        const SizedBox(height:8),
        Text('Items:', style: Theme.of(context).textTheme.titleMedium),
        ...order!.items.map((it)=> ListTile(title: Text(it.recipe.title), subtitle: Text('x${it.quantity} - \$${it.price}'))),
      ])),
    );
  }
}
