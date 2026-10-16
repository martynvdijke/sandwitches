import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:wakelock_plus/wakelock_plus.dart';
import '../../core/providers.dart';

class CookingScreen extends ConsumerStatefulWidget {
  final int recipeId;
  const CookingScreen({super.key, required this.recipeId});
  @override
  ConsumerState<CookingScreen> createState()=> _CookingState();
}
class _CookingState extends ConsumerState<CookingScreen> {
  final controller = PageController();
  int idx=0;
  @override
  void initState(){
    super.initState();
    WakelockPlus.enable();
  }
  @override
  void dispose(){
    WakelockPlus.disable();
    controller.dispose();
    super.dispose();
  }
  @override
  Widget build(BuildContext context){
    return FutureBuilder(
      future: ref.read(repositoryProvider).fetchRecipe(widget.recipeId),
      builder: (c,snap){
        if (snap.connectionState==ConnectionState.waiting) return Scaffold(appBar: AppBar(title: const Text('Cooking')), body: const Center(child: CircularProgressIndicator()));
        if (snap.hasError) return Scaffold(appBar: AppBar(title: const Text('Cooking')), body: Center(child: Text('Error ${snap.error}')));
        final r = snap.data as dynamic;
        final steps = (r.instructions as String).split('\n').where((e)=> e.trim().isNotEmpty).toList();
        if (steps.isEmpty) steps.add(r.instructions as String);
        return Scaffold(
          appBar: AppBar(title: Text('Cooking: ${r.title} - ${idx+1}/${steps.length}')),
          body: PageView.builder(
            controller: controller,
            onPageChanged: (i)=> setState(()=> idx=i),
            itemCount: steps.length,
            itemBuilder: (c,i)=> Padding(padding: const EdgeInsets.all(32), child: Center(child: Text(steps[i], style: Theme.of(context).textTheme.headlineSmall, textAlign: TextAlign.center))),
          ),
          bottomNavigationBar: Padding(padding: const EdgeInsets.all(16), child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children:[
            ElevatedButton(onPressed: idx>0? ()=> controller.previousPage(duration: const Duration(milliseconds:300), curve: Curves.ease): null, child: const Text('Prev')),
            Text('${idx+1} / ${steps.length}'),
            ElevatedButton(onPressed: idx<steps.length-1? ()=> controller.nextPage(duration: const Duration(milliseconds:300), curve: Curves.ease): null, child: const Text('Next')),
          ])),
        );
      },
    );
  }
}
