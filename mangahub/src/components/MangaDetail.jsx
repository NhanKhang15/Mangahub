import { motion } from 'motion/react';
import { Play, Bookmark, Star, SortAsc, ChevronRight, ArrowLeft } from 'lucide-react';

const MOCK_CHAPTERS = Array.from({ length: 15 }, (_, i) => ({
  id: String(15 - i),
  number: 15 - i,
  title: `The Path of Shadows Part ${15 - i}`,
  date: `${i + 1} day${i === 0 ? '' : 's'} ago`
}));

export function MangaDetail({ manga, onBack }) {
  return (
    <motion.div 
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-12"
    >
      <button onClick={onBack} className="flex items-center gap-2 text-on-surface-variant hover:text-primary transition-colors font-bold group">
        <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" /> Back to Discover
      </button>

      {/* Detail Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 blur-3xl opacity-20">
          <img src={manga.cover} className="w-full h-full object-cover" />
        </div>
        <div className="relative flex flex-col md:flex-row gap-8 items-start md:items-end p-6 bg-surface-container/30 rounded-3xl border border-white/20">
          <div className="shrink-0 w-48 md:w-64 aspect-[2/3] rounded-2xl overflow-hidden shadow-2xl">
            <img src={manga.cover} className="w-full h-full object-cover" />
          </div>
          <div className="flex-1 space-y-6">
            <div className="flex flex-wrap gap-2">
              <span className="px-3 py-1 bg-primary/10 text-primary rounded-full text-xs font-bold font-sans">Trending #1</span>
              {manga.tags.map(t => (
                <span key={t} className="px-3 py-1 bg-surface-container text-on-surface-variant rounded-full text-xs font-semibold">{t}</span>
              ))}
            </div>
            <h1 className="text-4xl md:text-5xl font-extrabold text-on-surface tracking-tight">{manga.title}</h1>
            <div className="flex items-center gap-6 text-on-surface-variant text-sm font-semibold">
              <span className="flex items-center gap-1">
                <Star className="w-4 h-4 text-amber-400 fill-amber-400" /> 4.9 (12k)
              </span>
              <span>Author: {manga.author}</span>
              <span>{manga.chapters} Chapters</span>
            </div>
          </div>
        </div>
      </section>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-12">
        <div className="lg:col-span-7 space-y-10">
          <div className="flex flex-wrap gap-4">
            <button className="bg-primary hover:bg-primary-container text-white px-8 py-4 rounded-xl font-bold flex items-center gap-2 shadow-lg shadow-primary/20 transition-all hover:scale-105 active:scale-95">
              <Play className="w-5 h-5 fill-current" /> Read Now
            </button>
            <button className="bg-surface-container hover:bg-surface-container-high text-on-surface px-8 py-4 rounded-xl font-bold flex items-center gap-2 transition-all active:scale-95">
              <Bookmark className="w-5 h-5" /> Subscribe
            </button>
          </div>

          <div className="p-8 bg-white rounded-3xl shadow-sm border border-surface-container">
            <h3 className="text-2xl font-bold mb-4">Synopsis</h3>
            <p className="text-lg text-on-surface-variant leading-relaxed font-medium">
              {manga.synopsis} {manga.synopsis}
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { label: 'Status', value: manga.status },
              { label: 'Released', value: '2021' },
              { label: 'Views', value: '4.2M' },
              { label: 'Type', value: 'Manga' }
            ].map(stat => (
              <div key={stat.label} className="p-6 bg-surface-container-low rounded-2xl text-center border border-surface-container/50">
                <span className="block text-on-surface-variant/50 text-[10px] font-bold uppercase tracking-widest mb-1">{stat.label}</span>
                <span className="font-bold text-on-surface">{stat.value}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="lg:col-span-5">
           <div className="bg-white rounded-3xl shadow-sm border border-surface-container flex flex-col h-[600px]">
             <div className="p-6 border-b border-surface-container flex justify-between items-center">
               <h3 className="text-xl font-bold">Chapters</h3>
               <button className="p-2 hover:bg-surface-container rounded-lg transition-colors"><SortAsc className="w-5 h-5 text-on-surface-variant" /></button>
             </div>
             <div className="flex-1 overflow-y-auto p-4 space-y-1">
               {MOCK_CHAPTERS.map((ch, i) => (
                 <button key={ch.id} className={`w-full flex items-center justify-between p-4 rounded-xl transition-all group ${i === 0 ? 'bg-primary/5 border border-primary/10' : 'hover:bg-surface-container-low'}`}>
                   <div className="flex items-center gap-4 text-left">
                     <div className={`w-10 h-10 rounded-lg flex items-center justify-center font-bold ${i === 0 ? 'bg-primary text-white' : 'bg-surface-container text-on-surface-variant'}`}>{ch.number}</div>
                     <div>
                       <span className="font-bold block text-sm line-clamp-1">{ch.title}</span>
                       <span className="text-[10px] font-bold text-on-surface-variant/50 uppercase tracking-tighter">{ch.date}</span>
                     </div>
                   </div>
                   <ChevronRight className={`w-5 h-5 transition-transform group-hover:translate-x-1 ${i === 0 ? 'text-primary' : 'text-on-surface-variant/30'}`} />
                 </button>
               ))}
             </div>
             <div className="p-6 border-t border-surface-container">
               <button className="w-full py-3 text-primary font-bold hover:bg-primary/5 rounded-xl transition-colors">View All {manga.chapters} Chapters</button>
             </div>
           </div>
        </div>
      </div>
    </motion.div>
  );
}
