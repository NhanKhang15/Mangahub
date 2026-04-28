import { motion } from 'motion/react';
import { Play, Plus, Star, ChevronRight, ChevronLeft } from 'lucide-react';
import { MANGA_DATA } from '../data';

export function MangaCard({ manga, onClick }) {
  return (
    <motion.div 
      whileHover={{ y: -8 }}
      onClick={onClick}
      className="flex-none w-48 group cursor-pointer"
    >
      <div className="relative aspect-[2/3] rounded-2xl overflow-hidden mb-3 shadow-sm group-hover:shadow-xl transition-all duration-300">
        <img src={manga.cover} alt={manga.title} className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500" />
        <div className="absolute top-2 right-2 bg-black/60 backdrop-blur-md text-white px-2 py-1 rounded-lg flex items-center gap-1">
          <Star className="w-3 h-3 text-yellow-400 fill-yellow-400" />
          <span className="text-[10px] font-bold">{manga.rating}</span>
        </div>
      </div>
      <h4 className="font-bold text-on-surface line-clamp-1 mb-1 group-hover:text-primary transition-colors">{manga.title}</h4>
      <p className="text-on-surface-variant text-xs">{manga.tags.slice(0, 2).join(', ')}</p>
    </motion.div>
  );
}

export function Home({ onSelectManga }) {
  const featured = MANGA_DATA[0];

  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <section className="relative rounded-3xl overflow-hidden h-[480px] group cursor-pointer" onClick={() => onSelectManga(featured)}>
        <img src={featured.cover} alt={featured.title} className="absolute inset-0 w-full h-full object-cover transition-transform duration-[2s] group-hover:scale-105" />
        <div className="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/40 to-transparent" />
        <div className="absolute bottom-0 left-0 p-8 md:p-12 w-full md:w-2/3">
          <div className="flex items-center gap-2 mb-4">
            <span className="px-3 py-1 bg-primary text-white text-[10px] font-bold rounded-full uppercase tracking-wider">Featured Series</span>
            <span className="px-3 py-1 bg-white/20 backdrop-blur-md text-white text-[10px] font-bold rounded-full uppercase tracking-wider">Updated Now</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-extrabold text-white mb-4 tracking-tight leading-tight">{featured.title}</h2>
          <p className="text-lg text-slate-200 mb-8 max-w-xl line-clamp-2">{featured.synopsis}</p>
          <div className="flex gap-4">
            <button className="bg-primary hover:bg-primary-container text-white px-8 py-4 rounded-xl transition-all active:scale-95 flex items-center gap-2 font-bold shadow-lg shadow-primary/20">
              <Play className="w-5 h-5 fill-current" /> Read Now
            </button>
            <button className="bg-white/10 hover:bg-white/20 backdrop-blur-md text-white px-8 py-4 rounded-xl transition-all active:scale-95 flex items-center gap-2 font-bold">
              <Plus className="w-5 h-5" /> Library
            </button>
          </div>
        </div>
      </section>

      {/* Trending Rail */}
      <section>
        <div className="flex justify-between items-end mb-6">
          <div>
            <h3 className="text-3xl font-bold text-on-surface tracking-tight">Trending Manga</h3>
            <p className="text-on-surface-variant mt-1">What the community is devouring right now</p>
          </div>
          <button className="text-primary font-bold hover:underline flex items-center gap-1 transition-all text-sm group">
            View All <ChevronRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
          </button>
        </div>
        <div className="flex gap-6 overflow-x-auto pb-6 hide-scrollbar -mx-6 px-6">
          {MANGA_DATA.map(m => (
            <MangaCard key={m.id} manga={m} onClick={() => onSelectManga(m)} />
          ))}
        </div>
      </section>

      {/* Recently Updated Grid */}
      <section>
        <div className="flex justify-between items-end mb-8">
          <div>
            <h3 className="text-3xl font-bold text-on-surface tracking-tight">Recently Updated</h3>
            <p className="text-on-surface-variant mt-1">Freshly baked chapters for your library</p>
          </div>
          <div className="flex gap-2">
            <button className="w-10 h-10 rounded-full border border-surface-container flex items-center justify-center hover:bg-white hover:shadow-md transition-all active:scale-90">
              <ChevronLeft className="w-5 h-5 text-on-surface-variant" />
            </button>
            <button className="w-10 h-10 rounded-full border border-surface-container flex items-center justify-center hover:bg-white hover:shadow-md transition-all active:scale-90">
              <ChevronRight className="w-5 h-5 text-on-surface-variant" />
            </button>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
           <div className="md:col-span-2 md:row-span-2 relative rounded-3xl overflow-hidden group cursor-pointer">
             <img src="https://images.unsplash.com/photo-1618336753974-aae8e04506aa?q=80&w=1000&auto=format&fit=crop" className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-700" />
             <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent" />
             <div className="absolute bottom-0 left-0 p-8">
               <span className="bg-primary text-white px-3 py-1 rounded-full text-xs font-bold mb-3 inline-block">Ch. 248 Just Out</span>
               <h4 className="text-2xl font-bold text-white">Chronicles of the Sky-Born</h4>
               <p className="text-slate-300 mt-2">The war for the floating isles reaches its breaking point.</p>
             </div>
           </div>
           {MANGA_DATA.slice(1, 5).map(m => (
             <div key={m.id} onClick={() => onSelectManga(m)} className="bg-white p-4 rounded-2xl flex gap-4 border border-surface-container hover:shadow-lg transition-all cursor-pointer group">
               <div className="w-16 h-20 rounded-lg overflow-hidden flex-shrink-0">
                 <img src={m.cover} className="w-full h-full object-cover group-hover:scale-110 transition-transform" />
               </div>
               <div className="flex flex-col justify-center overflow-hidden">
                 <h5 className="font-bold text-on-surface truncate">{m.title}</h5>
                 <p className="text-primary text-xs font-extrabold mt-1">Ch. {m.chapters}</p>
                 <p className="text-on-surface-variant text-[10px] mt-1">2 hours ago</p>
               </div>
             </div>
           ))}
        </div>
      </section>
    </div>
  );
}
