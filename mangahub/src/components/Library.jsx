import { motion } from 'motion/react';
import { Play, RotateCcw, Plus, Search } from 'lucide-react';
import { MANGA_DATA } from '../data';

export function Library({ onSelectManga }) {
  return (
    <div className="space-y-10">
      <header className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <div className="space-y-2">
          <h1 className="text-5xl font-extrabold text-on-surface tracking-tight">Your Library</h1>
          <p className="text-on-surface-variant max-w-xl font-medium">Manage your collection, track your progress, and jump right back into your stories.</p>
        </div>
        <div className="flex items-center gap-2 bg-surface-container p-1 rounded-xl">
           <button className="px-6 py-2 bg-white rounded-lg shadow-sm font-bold text-primary text-sm transition-all">All Manga</button>
           <button className="px-6 py-2 font-bold text-on-surface-variant hover:text-on-surface text-sm transition-all">Reading</button>
           <button className="px-6 py-2 font-bold text-on-surface-variant hover:text-on-surface text-sm transition-all">Finished</button>
        </div>
      </header>

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5 gap-8">
        {MANGA_DATA.map(m => (
          <motion.div 
            key={m.id}
            whileHover={{ y: -8 }}
            className="group bg-white rounded-2xl overflow-hidden border border-surface-container shadow-sm hover:shadow-xl transition-all duration-300"
          >
            <div className="relative aspect-[2/3] overflow-hidden">
              <img src={m.cover} className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110" />
              {m.status === 'Ongoing' && (
                <div className="absolute top-3 right-3 bg-white/90 backdrop-blur-md px-2 py-1 rounded-lg">
                  <span className="text-[10px] font-extrabold text-primary uppercase tracking-wider">{m.status}</span>
                </div>
              )}
            </div>
            <div className="p-5 space-y-4">
              <div>
                <h3 className="font-bold text-on-surface truncate group-hover:text-primary transition-colors">{m.title}</h3>
                <div className="flex items-center justify-between mt-1">
                  <span className="text-[10px] font-bold text-on-surface-variant/50 uppercase tracking-widest">Ch. 25/{m.chapters}</span>
                  <span className="text-[10px] font-bold text-primary">25%</span>
                </div>
              </div>
              <div className="w-full h-1.5 bg-surface-container rounded-full overflow-hidden">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: '25%' }}
                  className="h-full bg-primary" 
                />
              </div>
              <button 
                onClick={() => onSelectManga(m)}
                className="w-full py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-container transition-all flex items-center justify-center gap-2 active:scale-95"
              >
                <Play className="w-4 h-4 fill-current" /> Continue
              </button>
            </div>
          </motion.div>
        ))}

        <motion.div 
          whileHover={{ scale: 0.98 }}
          className="flex flex-col items-center justify-center aspect-[2/3] border-2 border-dashed border-surface-container rounded-2xl hover:bg-primary/5 hover:border-primary/20 transition-all cursor-pointer group"
        >
          <div className="w-12 h-12 bg-surface-container rounded-full flex items-center justify-center mb-3 group-hover:bg-primary-fixed transition-colors">
            <Plus className="w-6 h-6 text-primary" />
          </div>
          <span className="font-bold text-on-surface-variant text-sm group-hover:text-primary transition-colors">Add New Manga</span>
        </motion.div>
      </div>
    </div>
  );
}
