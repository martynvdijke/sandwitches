// ignore_for_file: use_build_context_synchronously
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:image_picker/image_picker.dart';
import '../../core/providers.dart';
import '../../core/models.dart';

class RecipeDetailScreen extends ConsumerStatefulWidget {
  final int recipeId;
  const RecipeDetailScreen({super.key, required this.recipeId});
  @override
  ConsumerState<RecipeDetailScreen> createState()=> _DetailState();
}
class _DetailState extends ConsumerState<RecipeDetailScreen> {
  int targetServings = 1;
  List<bool> checked = [];
  Recipe? recipe;

  @override
  Widget build(BuildContext context){
    final base = ref.watch(baseUrlProvider) ?? '';
    return Scaffold(
      appBar: AppBar(title: const Text('Recipe')),
      body: FutureBuilder<Recipe>(
        future: ref.read(repositoryProvider).fetchRecipe(widget.recipeId),
        builder: (c,snap){
          if (snap.connectionState==ConnectionState.waiting) return const Center(child: CircularProgressIndicator());
          if (snap.hasError) return Center(child: Text('Error: ${snap.error}'));
          if (!snap.hasData) return const Center(child: Text('Not found'));
          final r = snap.data!;
          if (recipe==null) { recipe=r; targetServings=r.servings; }
          final ingLines = r.ingredients.split('\n').where((e)=>e.trim().isNotEmpty).toList();
          return ListView(padding: const EdgeInsets.all(12), children:[
            if (r.image.isNotEmpty) CachedNetworkImage(imageUrl: buildThumbUrl(base, r.image, w:800), height:200, width: double.infinity, fit: BoxFit.cover, errorWidget: (c,s,e)=> Container(color: Colors.grey[300], height:200, child: const Icon(Icons.broken_image))),
            const SizedBox(height:8),
            Text(r.title, style: Theme.of(context).textTheme.headlineSmall),
            Text(r.description),
            const SizedBox(height:8),
            Wrap(spacing:6, children: r.tags.map((t)=> Chip(label: Text(t.name))).toList()),
            const SizedBox(height:8),
            Text('Servings: $targetServings  Price: \$${r.price}'),
            Row(children:[
              IconButton(icon: const Icon(Icons.remove), onPressed: ()=> setState(()=> targetServings = targetServings>1? targetServings-1:1)),
              Text('$targetServings'),
              IconButton(icon: const Icon(Icons.add), onPressed: ()=> setState(()=> targetServings++)),
            ]),
            FutureBuilder<List<ScaledIngredient>>(
              future: ref.read(repositoryProvider).scaleIngredients(r.id, targetServings),
              builder: (c,ss){
                if (ss.connectionState==ConnectionState.waiting) return const LinearProgressIndicator();
                if (ss.hasError) return Text('Scale error: ${ss.error}');
                final list = ss.data ?? [];
                if (list.isEmpty) {
                  if (checked.length != ingLines.length) checked = List.filled(ingLines.length, false);
                  return Column(children: List.generate(ingLines.length, (i)=> CheckboxListTile(value: checked.length>i? checked[i]:false, onChanged: (v)=> setState(()=> checked[i]=v!), title: Text(ingLines[i]))));
                }
                if (checked.length != list.length) checked = List.filled(list.length, false);
                return Column(children: List.generate(list.length, (i)=> CheckboxListTile(value: checked[i], onChanged: (v)=> setState(()=> checked[i]=v!), title: Text(list[i].scaledLine.isNotEmpty? list[i].scaledLine : list[i].originalLine), subtitle: Text('${list[i].quantity?.toString() ?? ''} ${list[i].unit}'))));
              },
            ),
            const SizedBox(height:8),
            Text('Instructions:', style: Theme.of(context).textTheme.titleMedium),
            Text(r.instructions),
            const SizedBox(height:12),
            ElevatedButton.icon(icon: const Icon(Icons.soup_kitchen), label: const Text('Cooking Mode'), onPressed: ()=> context.push('/recipe/${r.id}/cooking')),
            const SizedBox(height:8),
            Row(children:[
              Expanded(child: ElevatedButton(onPressed: () async {
                try { final repo = ref.read(repositoryProvider); await repo.addToCart(r.id, targetServings); if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Added to cart'))); } catch(e){ if(mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Add to cart failed: $e'))); }
              }, child: const Text('Add to cart'))),
              const SizedBox(width:8),
              Expanded(child: OutlinedButton(onPressed: () async {
                try { final o = await ref.read(repositoryProvider).createOrder(r.id); if (mounted) context.push('/orders/track/${o.trackingToken.isNotEmpty? o.trackingToken : o.id}'); } catch(e){ if(mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Order failed: $e'))); }
              }, child: const Text('Order now'))),
            ]),
            const SizedBox(height:12),
            _RatingsSection(recipeId: r.id),
          ]);
        },
      ),
    );
  }
}

class _RatingsSection extends ConsumerStatefulWidget {
  final int recipeId;
  const _RatingsSection({required this.recipeId});
  @override
  ConsumerState<_RatingsSection> createState()=> _RatingsState();
}
class _RatingsState extends ConsumerState<_RatingsSection> {
  int refresh = 0;
  @override
  Widget build(BuildContext context){
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children:[
      FutureBuilder<Map<String,dynamic>>(
        future: ref.read(repositoryProvider).fetchRatingSummary(widget.recipeId),
        builder: (c,s){
          if (!s.hasData) return const SizedBox.shrink();
          final m = s.data!;
          return Text('Average: ${m['average']}  Count: ${m['count']}', style: Theme.of(context).textTheme.titleSmall);
        },
      ),
      FutureBuilder<List<Rating>>(
        future: ref.read(repositoryProvider).fetchRatings(widget.recipeId),
        builder: (c,s){
          if (s.connectionState==ConnectionState.waiting) return const LinearProgressIndicator();
          if (s.hasError) return Text('Ratings error: ${s.error}');
          final list = s.data ?? [];
          if (list.isEmpty) return const Text('No reviews yet');
          return Column(children: list.map((ra)=> ListTile(
            leading: Icon(Icons.star, color: Colors.amber.shade700),
            title: Text('${ra.score} stars - ${ra.comment}'),
            subtitle: Text(ra.user?['username']?.toString() ?? ''),
          )).toList());
        },
      ),
      const SizedBox(height:8),
      ElevatedButton(onPressed: ()=> showModalBottomSheet(context: context, isScrollControlled: true, builder: (_)=> RatingSheet(recipeId: widget.recipeId, onSubmitted: ()=> setState(()=> refresh++))), child: const Text('Write a review')),
    ]);
  }
}

class RatingSheet extends StatefulWidget {
  final int recipeId;
  final VoidCallback onSubmitted;
  const RatingSheet({super.key, required this.recipeId, required this.onSubmitted});
  @override
  State<RatingSheet> createState()=> _RatingSheetState();
}
class _RatingSheetState extends State<RatingSheet> {
  int score=5;
  final commentCtrl = TextEditingController();
  String? imagePath;
  bool loading=false;
  final ImagePicker _picker = ImagePicker();
  @override
  Widget build(BuildContext context){
    return Consumer(builder: (c,ref,_)=> Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom, left:16,right:16,top:16),
      child: Column(mainAxisSize: MainAxisSize.min, children:[
        Row(children: List.generate(5, (i)=> IconButton(icon: Icon(Icons.star, color: i<score? Colors.amber: Colors.grey), onPressed: ()=> setState(()=> score=i+1)))),
        TextField(controller: commentCtrl, decoration: const InputDecoration(labelText: 'Comment', border: OutlineInputBorder()), maxLines:3),
        const SizedBox(height:8),
        Row(children:[
          ElevatedButton.icon(icon: const Icon(Icons.photo), label: Text(imagePath==null? 'Pick photo':'Photo selected'), onPressed: () async {
            final x = await _picker.pickImage(source: ImageSource.gallery);
            if (x!=null) setState(()=> imagePath=x.path);
          }),
          if (imagePath!=null) IconButton(icon: const Icon(Icons.clear), onPressed: ()=> setState(()=> imagePath=null)),
        ]),
        const SizedBox(height:8),
        ElevatedButton(onPressed: loading? null: () async {
          setState(()=> loading=true);
          try {
            await ref.read(repositoryProvider).postRating(widget.recipeId, score, commentCtrl.text, imagePath: imagePath);
            widget.onSubmitted();
            if (context.mounted) Navigator.pop(context);
            if (context.mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Review submitted')));
          } catch(e){
            if (context.mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
          } finally { if(mounted) setState(()=> loading=false); }
        }, child: loading? const SizedBox(height:20,width:20,child:CircularProgressIndicator(strokeWidth:2)): const Text('Submit')),
        const SizedBox(height:16),
      ]),
    ));
  }
}
