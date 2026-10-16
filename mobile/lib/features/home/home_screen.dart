import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../core/providers.dart';
import '../../core/models.dart';
import '../../theme/theme_provider.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});
  @override
  ConsumerState<HomeScreen> createState()=> _HomeState();
}
class _HomeState extends ConsumerState<HomeScreen> {
  String search='';
  String selectedTag='';
  final searchCtrl = TextEditingController();

  @override
  Widget build(BuildContext context){
    final tagsAsync = ref.watch(tagsProvider);
    final rodAsync = ref.watch(recipeOfDayProvider);
    final recipesAsync = ref.watch(recipesProvider((search: search, tag: selectedTag)));
    final themeMode = ref.watch(themeNotifierProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Sandwitches'), actions: [
        PopupMenuButton<String>(onSelected: (v)=> ref.read(themeNotifierProvider.notifier).setTheme(v), itemBuilder: (_)=> const [
          PopupMenuItem(value:'system', child: Text('System')),
          PopupMenuItem(value:'light', child: Text('Light')),
          PopupMenuItem(value:'dark', child: Text('Dark')),
          PopupMenuItem(value:'eink', child: Text('E-Ink')),
        ], child: Padding(padding: const EdgeInsets.all(8), child: Row(children:[const Icon(Icons.palette), Text(' $themeMode')]))),
        IconButton(icon: const Icon(Icons.logout), onPressed: () async {
          await ref.read(storageProvider).clearAuth();
          await ref.read(persistCookieJarProvider).deleteAll();
          if (context.mounted) context.go('/login');
        }),
      ]),
      body: RefreshIndicator(onRefresh: () async {
        ref.invalidate(tagsProvider);
        ref.invalidate(recipeOfDayProvider);
        ref.invalidate(recipesProvider);
      }, child: ListView(padding: const EdgeInsets.all(12), children: [
        // Recipe of the Day
        rodAsync.when(
          data: (r){
            if (r==null) return const SizedBox.shrink();
            return Card(child: InkWell(onTap: ()=> context.push('/recipe/${r.id}'), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children:[
              if (r.image.isNotEmpty) _thumb(r),
              ListTile(title: Text('Recipe of the Day: ${r.title}'), subtitle: Text(r.description, maxLines:2, overflow: TextOverflow.ellipsis)),
            ])));
          },
          loading: ()=> const LinearProgressIndicator(),
          error: (e,s)=> Text('Failed to load ROTD: $e'),
        ),
        const SizedBox(height:12),
        TextField(controller: searchCtrl, decoration: InputDecoration(labelText: 'Search', suffixIcon: IconButton(icon: const Icon(Icons.search), onPressed: ()=> setState(()=> search=searchCtrl.text)), border: const OutlineInputBorder()), onSubmitted: (v)=> setState(()=> search=v)),
        const SizedBox(height:8),
        tagsAsync.when(
          data: (tags)=> Wrap(spacing:6, children:[
            ChoiceChip(label: const Text('All'), selected: selectedTag.isEmpty, onSelected: (_)=> setState(()=> selectedTag='')),
            for (final t in tags) ChoiceChip(label: Text(t.name), selected: selectedTag==t.slug, onSelected: (_)=> setState(()=> selectedTag = selectedTag==t.slug? '' : t.slug)),
          ]),
          loading: ()=> const SizedBox(height:20, child: LinearProgressIndicator()),
          error: (e,s)=> Text('Tags error: $e'),
        ),
        const SizedBox(height:12),
        recipesAsync.when(
          data: (list){
            if (list.isEmpty) return const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('No recipes found')));
            return GridView.builder(
              shrinkWrap: true, physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(crossAxisCount: 2, childAspectRatio: 0.75, crossAxisSpacing:8, mainAxisSpacing:8),
              itemCount: list.length,
              itemBuilder: (c,i){
                final r = list[i];
                return Card(clipBehavior: Clip.antiAlias, child: InkWell(onTap: ()=> context.push('/recipe/${r.id}'), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children:[
                  Expanded(child: r.image.isEmpty? Container(color: Colors.grey[300], child: const Center(child: Icon(Icons.image))): _thumb(r)),
                  Padding(padding: const EdgeInsets.all(8), child: Text(r.title, maxLines:2, overflow: TextOverflow.ellipsis, style: Theme.of(context).textTheme.titleSmall)),
                  Padding(padding: const EdgeInsets.symmetric(horizontal:8), child: Text('\$${r.price}', style: Theme.of(context).textTheme.bodySmall)),
                  const SizedBox(height:4),
                ])));
              },
            );
          },
          loading: ()=> const Center(child: CircularProgressIndicator()),
          error: (e,s)=> Text('Recipes error: $e'),
        ),
      ])),
    );
  }
  Widget _thumb(Recipe r){
    final base = ref.watch(baseUrlProvider) ?? '';
    final url = buildThumbUrl(base, r.image, w:400);
    if (url.isEmpty) return Container(color: Colors.grey[300], height:120);
    return CachedNetworkImage(imageUrl: url, height: 120, width: double.infinity, fit: BoxFit.cover, placeholder: (c,s)=> Container(color: Colors.grey[300]), errorWidget: (c,s,e)=> Container(color: Colors.grey[300], child: const Icon(Icons.broken_image)));
  }
}
